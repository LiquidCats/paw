package parser

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// Known 4-byte function selectors for ETH-related calls.
const (
	// SelectorWETHDeposit is keccak256("deposit()")[0:4] = 0xd0e30db0.
	SelectorWETHDeposit = "d0e30db0"
	// SelectorWETHWithdraw is keccak256("withdraw(uint256)")[0:4] = 0x2e1a7d4d.
	SelectorWETHWithdraw = "2e1a7d4d"
	// SelectorMulticall is keccak256("multicall(bytes[])")[0:4] = 0xac9650d8.
	// Used by Uniswap V3 and other protocols.
	SelectorMulticall = "ac9650d8"
	// SelectorMulticallDeadline is keccak256("multicall(uint256,bytes[])")[0:4] = 0x5ae401dc.
	// Timed variant used by Uniswap V3 NonfungiblePositionManager.
	SelectorMulticallDeadline = "5ae401dc"
	// SelectorMultiSend is keccak256("multiSend(bytes)")[0:4] = 0x8d80ff0a.
	// Used by Gnosis Safe.
	SelectorMultiSend = "8d80ff0a"
	// SelectorMultiSendCallOnly is keccak256("multiSendCallOnly(bytes)")[0:4] = 0x82ad56cb.
	// Call-only variant used by Gnosis Safe.
	SelectorMultiSendCallOnly = "82ad56cb"
)

// selectorLen is the byte length of a function selector (4 bytes = 8 hex chars).
const selectorLen = 4

// uint256Len is the byte length of a single ABI-encoded uint256 word (32 bytes).
const uint256Len = 32 //nolint:mnd

// Kind describes the type of ETH transfer detected from transaction input.
type Kind string

const (
	// KindNative is a plain ETH transfer with empty or "0x" input.
	KindNative Kind = "native"
	// KindWETHDeposit is a WETH deposit() call — ETH → WETH.
	KindWETHDeposit Kind = "weth_deposit"
	// KindWETHWithdraw is a WETH withdraw(uint256) call — WETH → ETH.
	KindWETHWithdraw Kind = "weth_withdraw"
	// KindMulticall is a batched multicall(bytes[]) containing sub-transfers.
	// SubCalls holds the parsed result for each individual sub-call.
	KindMulticall Kind = "multicall"
	// KindMultiSend is a Gnosis Safe multiSend/multiSendCallOnly call containing
	// packed sub-transactions, each with its own ETH value.
	// SubCalls holds the parsed result for each individual sub-transaction.
	KindMultiSend Kind = "multisend"
	// KindContractCall is an unknown contract call that carries ETH value.
	KindContractCall Kind = "contract_call"
	// KindNone means no ETH is being transferred (zero value, unrecognised input).
	KindNone Kind = ""
)

// Result holds the parsed transfer information derived from transaction input.
type Result struct {
	// Kind is the detected transfer type.
	Kind Kind
	// To is the destination address as a checksummed hex string (e.g. "0x1234…").
	// Populated for multiSend sub-transactions where the recipient is encoded in
	// the packed calldata.  Empty for all other kinds.
	To string
	// Amount is the ETH amount in wei when it can be decoded from the calldata
	// (e.g. WETH withdraw argument, or a multiSend sub-transaction value).
	// For native and deposit transfers the caller should use the transaction's
	// value field instead; Amount will be nil.
	Amount *big.Int
	// SubCalls contains the recursively parsed results for each sub-call inside
	// a multicall or multiSend transaction. Empty for all other kinds.
	SubCalls []Result
}

// Parse inspects the hex-encoded transaction input and the transaction value to
// determine whether the transaction represents an ETH transfer and, if so, what
// kind.  input may be an empty string, "0x", or a full hex calldata string.
// value is the transaction's native-ETH value in wei; pass nil or zero when the
// value is not available.
//
// For multicall and multiSend transactions, SubCalls in the returned Result
// contains the individually parsed sub-transactions.
func Parse(input string, value *big.Int) Result {
	raw := stripHexPrefix(input)

	if len(raw) < selectorLen*2 {
		return resultForNativeOrNone(value)
	}

	sel := strings.ToLower(raw[:selectorLen*2])
	body := raw[selectorLen*2:]

	switch sel {
	case SelectorWETHDeposit:
		// deposit() carries ETH in the value field; no calldata to decode.
		if len(body) == 0 {
			return Result{Kind: KindWETHDeposit}
		}
		// Extra data present — treat as deposit anyway; the value field still
		// conveys the ETH amount.
		return Result{Kind: KindWETHDeposit}

	case SelectorWETHWithdraw:
		return parseWETHWithdraw(body)

	case SelectorMulticall:
		// multicall(bytes[] data): ABI head is the pointer to bytes[].
		// Head starts at body offset 0.
		return parseMulticall(body, 0)

	case SelectorMulticallDeadline:
		// multicall(uint256 deadline, bytes[] data): skip the 32-byte deadline,
		// then the pointer to bytes[] is at body offset 32.
		return parseMulticall(body, wordSize)

	case SelectorMultiSend, SelectorMultiSendCallOnly:
		return parseMultiSend(body)

	default:
		if isPositive(value) {
			return Result{Kind: KindContractCall}
		}

		return Result{Kind: KindNone}
	}
}

// parseWETHWithdraw decodes the withdraw(uint256 wad) argument from the body
// (calldata after selector).
func parseWETHWithdraw(body string) Result {
	expectedLen := uint256Len * 2
	if len(body) < expectedLen {
		return Result{Kind: KindContractCall}
	}

	wordBytes, err := hex.DecodeString(body[:expectedLen])
	if err != nil {
		return Result{Kind: KindContractCall}
	}

	return Result{Kind: KindWETHWithdraw, Amount: new(big.Int).SetBytes(wordBytes)}
}

// parseMulticall decodes a multicall(bytes[] data) body.
// arrayHeadOffset is the byte offset within the decoded body where the bytes[]
// head pointer is located (0 for the basic variant, 32 for the deadline variant).
func parseMulticall(body string, arrayHeadOffset int) Result {
	raw, err := hex.DecodeString(body)
	if err != nil {
		return Result{Kind: KindContractCall}
	}

	calls, err := decodeBytesArrayParam(raw, arrayHeadOffset)
	if err != nil {
		return Result{Kind: KindContractCall}
	}

	subCalls := make([]Result, 0, len(calls))

	for _, callData := range calls {
		subInput := "0x" + hex.EncodeToString(callData)
		// Individual multicall sub-calls share the outer tx's ETH value and we
		// cannot split it per sub-call from calldata alone, so pass nil.
		subCalls = append(subCalls, Parse(subInput, nil))
	}

	return Result{Kind: KindMulticall, SubCalls: subCalls}
}

// parseMultiSend decodes a multiSend(bytes transactions) body.
func parseMultiSend(body string) Result {
	raw, err := hex.DecodeString(body)
	if err != nil {
		return Result{Kind: KindContractCall}
	}

	packed, err := decodeBytesParam(raw, 0)
	if err != nil {
		return Result{Kind: KindContractCall}
	}

	entries, err := decodeMultiSendTransactions(packed)
	if err != nil {
		return Result{Kind: KindContractCall}
	}

	subCalls := make([]Result, 0, len(entries))

	for _, entry := range entries {
		subInput := "0x" + hex.EncodeToString(entry.Data)
		sub := Parse(subInput, entry.Value)
		// Carry the decoded value for sub-calls that would otherwise rely on
		// the outer tx value (native and contract-call kinds).
		if sub.Amount == nil && isPositive(entry.Value) {
			sub.Amount = entry.Value
		}
		sub.To = toHexAddress(entry.To)
		subCalls = append(subCalls, sub)
	}

	return Result{Kind: KindMultiSend, SubCalls: subCalls}
}

// resultForNativeOrNone returns KindNative when value > 0, otherwise KindNone.
func resultForNativeOrNone(value *big.Int) Result {
	if isPositive(value) {
		return Result{Kind: KindNative}
	}

	return Result{Kind: KindNone}
}

// stripHexPrefix removes a leading "0x" or "0X" prefix and returns the
// remaining string.
func stripHexPrefix(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return s[2:]
	}

	return s
}

// isPositive returns true when v is non-nil and greater than zero.
func isPositive(v *big.Int) bool {
	return v != nil && v.Sign() > 0
}

// toHexAddress formats a 20-byte address as a lowercase hex string with 0x prefix.
func toHexAddress(addr [20]byte) string {
	return fmt.Sprintf("0x%s", hex.EncodeToString(addr[:]))
}
