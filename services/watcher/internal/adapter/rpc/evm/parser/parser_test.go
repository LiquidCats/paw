package parser_test

import (
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/LiquidCats/paw/services/watcher/internal/adapter/rpc/evm/parser"
)

func TestParse(t *testing.T) {
	t.Parallel()

	oneETH, _ := new(big.Int).SetString("1000000000000000000", 10)
	withdrawAmount, _ := new(big.Int).SetString("500000000000000000", 10) // 0.5 ETH

	tests := []struct {
		name          string
		input         string
		value         *big.Int
		wantKind      parser.Kind
		wantAmount    *big.Int
		wantSubCount  int
		checkSubCalls func(t *testing.T, subs []parser.Result)
	}{
		{
			name:     "empty input with value → native transfer",
			input:    "",
			value:    oneETH,
			wantKind: parser.KindNative,
		},
		{
			name:     "0x input with value → native transfer",
			input:    "0x",
			value:    oneETH,
			wantKind: parser.KindNative,
		},
		{
			name:     "empty input zero value → no transfer",
			input:    "",
			value:    big.NewInt(0),
			wantKind: parser.KindNone,
		},
		{
			name:     "empty input nil value → no transfer",
			input:    "",
			value:    nil,
			wantKind: parser.KindNone,
		},
		{
			name:     "WETH deposit with 0x prefix",
			input:    "0xd0e30db0",
			value:    oneETH,
			wantKind: parser.KindWETHDeposit,
		},
		{
			name:     "WETH deposit uppercase selector",
			input:    "0xD0E30DB0",
			value:    oneETH,
			wantKind: parser.KindWETHDeposit,
		},
		{
			name:       "WETH withdraw with amount",
			input:      buildWETHWithdrawInput(withdrawAmount),
			value:      nil,
			wantKind:   parser.KindWETHWithdraw,
			wantAmount: withdrawAmount,
		},
		{
			name:     "WETH withdraw malformed calldata → contract call",
			input:    "0x2e1a7d4d1234",
			value:    nil,
			wantKind: parser.KindContractCall,
		},
		{
			name:     "unknown calldata with ETH value → contract call",
			input:    "0xa9059cbb000000000000000000000000deadbeef",
			value:    oneETH,
			wantKind: parser.KindContractCall,
		},
		{
			name:     "unknown calldata zero value → no transfer",
			input:    "0xa9059cbb000000000000000000000000deadbeef",
			value:    big.NewInt(0),
			wantKind: parser.KindNone,
		},
		// multicall(bytes[]) — two sub-calls: WETH deposit + WETH withdraw.
		{
			name:         "multicall with WETH deposit and withdraw sub-calls",
			input:        buildMulticallInput([]string{"0xd0e30db0", buildWETHWithdrawInput(withdrawAmount)}),
			value:        oneETH,
			wantKind:     parser.KindMulticall,
			wantSubCount: 2,
			checkSubCalls: func(t *testing.T, subs []parser.Result) {
				t.Helper()
				if subs[0].Kind != parser.KindWETHDeposit {
					t.Errorf("sub[0].Kind = %q, want %q", subs[0].Kind, parser.KindWETHDeposit)
				}
				if subs[1].Kind != parser.KindWETHWithdraw {
					t.Errorf("sub[1].Kind = %q, want %q", subs[1].Kind, parser.KindWETHWithdraw)
				}
				if subs[1].Amount == nil || subs[1].Amount.Cmp(withdrawAmount) != 0 {
					t.Errorf("sub[1].Amount = %v, want %s", subs[1].Amount, withdrawAmount)
				}
			},
		},
		// multicall(uint256,bytes[]) — deadline variant with one native sub-call.
		{
			name:         "multicall with deadline wrapping a native ETH sub-call",
			input:        buildMulticallDeadlineInput(12345, []string{"0x"}),
			value:        oneETH,
			wantKind:     parser.KindMulticall,
			wantSubCount: 1,
			checkSubCalls: func(t *testing.T, subs []parser.Result) {
				t.Helper()
				// Sub-call has no value info from calldata, so kind is KindNone.
				if subs[0].Kind != parser.KindNone {
					t.Errorf("sub[0].Kind = %q, want %q", subs[0].Kind, parser.KindNone)
				}
			},
		},
		// multiSend — two packed sub-transactions with ETH values and distinct addresses.
		{
			name: "multiSend with two ETH sub-transactions",
			input: buildMultiSendInput([]multiSendEntry{
				{to: addrBytes("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"), value: oneETH, data: nil},
				{to: addrBytes("0xcafecafecafecafecafecafecafecafecafecafe"), value: withdrawAmount, data: hexBytes("0x")},
			}),
			value:        nil,
			wantKind:     parser.KindMultiSend,
			wantSubCount: 2,
			checkSubCalls: func(t *testing.T, subs []parser.Result) {
				t.Helper()
				for i, sub := range subs {
					if sub.Kind != parser.KindNative {
						t.Errorf("sub[%d].Kind = %q, want %q", i, sub.Kind, parser.KindNative)
					}
				}
				if subs[0].Amount == nil || subs[0].Amount.Cmp(oneETH) != 0 {
					t.Errorf("sub[0].Amount = %v, want %s", subs[0].Amount, oneETH)
				}
				if subs[0].To != "0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef" {
					t.Errorf("sub[0].To = %q, want 0xdeadbeef...", subs[0].To)
				}
				if subs[1].Amount == nil || subs[1].Amount.Cmp(withdrawAmount) != 0 {
					t.Errorf("sub[1].Amount = %v, want %s", subs[1].Amount, withdrawAmount)
				}
				if subs[1].To != "0xcafecafecafecafecafecafecafecafecafecafe" {
					t.Errorf("sub[1].To = %q, want 0xcafecafe...", subs[1].To)
				}
			},
		},
		// multiSend — sub-transaction containing a WETH deposit.
		{
			name: "multiSend with WETH deposit sub-transaction",
			input: buildMultiSendInput([]multiSendEntry{
				{to: addrBytes("0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"), value: oneETH, data: hexBytes("0xd0e30db0")},
			}),
			value:        nil,
			wantKind:     parser.KindMultiSend,
			wantSubCount: 1,
			checkSubCalls: func(t *testing.T, subs []parser.Result) {
				t.Helper()
				if subs[0].Kind != parser.KindWETHDeposit {
					t.Errorf("sub[0].Kind = %q, want %q", subs[0].Kind, parser.KindWETHDeposit)
				}
				if subs[0].To != "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2" {
					t.Errorf("sub[0].To = %q, want WETH contract address", subs[0].To)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := parser.Parse(tc.input, tc.value)

			if got.Kind != tc.wantKind {
				t.Errorf("Kind = %q, want %q", got.Kind, tc.wantKind)
			}

			if tc.wantAmount != nil {
				if got.Amount == nil {
					t.Fatalf("Amount = nil, want %s", tc.wantAmount)
				}

				if got.Amount.Cmp(tc.wantAmount) != 0 {
					t.Errorf("Amount = %s, want %s", got.Amount, tc.wantAmount)
				}
			}

			if tc.wantSubCount > 0 {
				if len(got.SubCalls) != tc.wantSubCount {
					t.Fatalf("len(SubCalls) = %d, want %d", len(got.SubCalls), tc.wantSubCount)
				}
			}

			if tc.checkSubCalls != nil {
				tc.checkSubCalls(t, got.SubCalls)
			}
		})
	}
}

// ---- builder helpers --------------------------------------------------------

// buildWETHWithdrawInput builds hex calldata for withdraw(uint256 wad).
func buildWETHWithdrawInput(wad *big.Int) string {
	padded := abiPadBigInt(wad)
	return "0x" + "2e1a7d4d" + hex.EncodeToString(padded)
}

// buildMulticallInput builds calldata for multicall(bytes[] data).
func buildMulticallInput(calls []string) string {
	return "0x" + "ac9650d8" + encodeBytesArray(calls)
}

// buildMulticallDeadlineInput builds calldata for multicall(uint256,bytes[]).
func buildMulticallDeadlineInput(deadline uint64, calls []string) string {
	sel := "5ae401dc"
	deadlineWord := make([]byte, 32) //nolint:mnd
	binary.BigEndian.PutUint64(deadlineWord[24:], deadline)

	// The bytes[] pointer is 0x40 (64) since there are 2 head words (deadline + pointer).
	arrayData := encodeBytesArrayRaw(calls)
	pointer := abiPadUint64(64) //nolint:mnd

	return "0x" + sel + hex.EncodeToString(deadlineWord) + hex.EncodeToString(pointer) + hex.EncodeToString(arrayData)
}

// multiSendEntry is a helper for building multiSend calldata in tests.
type multiSendEntry struct {
	to    [20]byte
	value *big.Int
	data  []byte
}

// buildMultiSendInput builds calldata for multiSend(bytes transactions).
func buildMultiSendInput(entries []multiSendEntry) string {
	packed := packMultiSendTransactions(entries)

	// ABI-encode: pointer(0x20) + length + data (padded to 32).
	pointer := abiPadUint64(32) //nolint:mnd
	length := abiPadUint64(uint64(len(packed)))

	padded := packed
	if rem := len(packed) % 32; rem != 0 { //nolint:mnd
		padded = append(padded, make([]byte, 32-rem)...) //nolint:mnd
	}

	body := append(pointer, length...)
	body = append(body, padded...)

	return "0x" + "8d80ff0a" + hex.EncodeToString(body)
}

// packMultiSendTransactions packs entries into the Gnosis multiSend format:
// { operation(1) | to(20) | value(32) | dataLength(32) | data }.
func packMultiSendTransactions(entries []multiSendEntry) []byte {
	var buf []byte

	for _, e := range entries {
		buf = append(buf, 0x00)          // operation = CALL
		buf = append(buf, e.to[:]...)   // to address (20 bytes)
		buf = append(buf, abiPadBigInt(e.value)...)
		buf = append(buf, abiPadUint64(uint64(len(e.data)))...)
		buf = append(buf, e.data...)
	}

	return buf
}

// encodeBytesArray ABI-encodes a bytes[] starting with the array length pointer
// (0x20) then the array content.
func encodeBytesArray(calls []string) string {
	pointer := abiPadUint64(32) //nolint:mnd
	return hex.EncodeToString(pointer) + hex.EncodeToString(encodeBytesArrayRaw(calls))
}

// encodeBytesArrayRaw ABI-encodes the array content (length word + element
// offsets + element data) without the outer head pointer.
func encodeBytesArrayRaw(calls []string) []byte {
	n := len(calls)
	elems := make([][]byte, n)

	for i, c := range calls {
		elems[i] = hexBytes(c)
	}

	// Layout: [length] [offset_0 .. offset_{n-1}] [elem_0 .. elem_{n-1}]
	// Each element: [length word] [data padded to 32]
	// Offsets are relative to the start of this array encoding (the length word).

	// Compute element offsets (relative to start of array encoding = 0).
	// Elements start after: 1 (length) + n (offsets) = (n+1) words.
	offsets := make([]uint64, n)
	cursor := uint64((n + 1) * 32) //nolint:mnd

	for i, elem := range elems {
		offsets[i] = cursor
		cursor += 32 + uint64(padLen(len(elem))) //nolint:mnd
	}

	var buf []byte
	buf = append(buf, abiPadUint64(uint64(n))...)

	for _, off := range offsets {
		buf = append(buf, abiPadUint64(off)...)
	}

	for _, elem := range elems {
		buf = append(buf, abiPadUint64(uint64(len(elem)))...)
		padded := make([]byte, padLen(len(elem)))
		copy(padded, elem)
		buf = append(buf, padded...)
	}

	return buf
}

// ---- misc helpers -----------------------------------------------------------

func abiPadBigInt(v *big.Int) []byte {
	if v == nil {
		return make([]byte, 32) //nolint:mnd
	}

	b := v.Bytes()
	padded := make([]byte, 32) //nolint:mnd
	copy(padded[32-len(b):], b)

	return padded
}

func abiPadUint64(v uint64) []byte {
	buf := make([]byte, 32) //nolint:mnd
	binary.BigEndian.PutUint64(buf[24:], v)

	return buf
}

// padLen returns the byte length of v rounded up to the nearest multiple of 32.
func padLen(v int) int {
	if v == 0 {
		return 0
	}

	return ((v + 31) / 32) * 32 //nolint:mnd
}

// addrBytes decodes a 20-byte Ethereum address from a hex string into [20]byte.
func addrBytes(s string) [20]byte {
	b := hexBytes(s)

	var addr [20]byte
	copy(addr[20-len(b):], b)

	return addr
}

// hexBytes decodes a hex string (with or without 0x prefix) into bytes.
func hexBytes(s string) []byte {
	if s == "" || s == "0x" {
		return []byte{}
	}

	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		s = s[2:]
	}

	b, _ := hex.DecodeString(s)

	return b
}
