package parser

import (
	"encoding/binary"
	"fmt"
	"math/big"
)

// wordSize is the size of an ABI-encoded word in bytes.
const wordSize = 32 //nolint:mnd

// readWord returns the 32-byte word at the given byte offset inside raw.
func readWord(raw []byte, offset int) ([]byte, error) {
	end := offset + wordSize
	if len(raw) < end {
		return nil, fmt.Errorf("abi: read word at offset %d: need %d bytes, have %d", offset, end, len(raw))
	}

	return raw[offset:end], nil
}

// readUint64Word reads a 32-byte word at offset and returns its value as uint64.
// Values that overflow uint64 are rejected.
func readUint64Word(raw []byte, offset int) (uint64, error) {
	word, err := readWord(raw, offset)
	if err != nil {
		return 0, err
	}

	// Only the last 8 bytes should be non-zero for a sane uint64.
	for _, b := range word[:wordSize-8] { //nolint:mnd
		if b != 0 {
			return 0, fmt.Errorf("abi: uint64 overflow at offset %d", offset)
		}
	}

	return binary.BigEndian.Uint64(word[wordSize-8:]), nil //nolint:mnd
}

// decodeBytesParam decodes a single ABI `bytes` parameter whose head pointer
// resides at headOffset inside raw.  Returns the raw byte slice.
func decodeBytesParam(raw []byte, headOffset int) ([]byte, error) {
	// Read the pointer (relative to the start of raw).
	dataOffset, err := readUint64Word(raw, headOffset)
	if err != nil {
		return nil, fmt.Errorf("abi: bytes pointer: %w", err)
	}

	// Read the length word at the pointed-to location.
	length, err := readUint64Word(raw, int(dataOffset))
	if err != nil {
		return nil, fmt.Errorf("abi: bytes length: %w", err)
	}

	dataStart := int(dataOffset) + wordSize
	dataEnd := dataStart + int(length)

	if len(raw) < dataEnd {
		return nil, fmt.Errorf("abi: bytes data truncated: need %d, have %d", dataEnd, len(raw))
	}

	out := make([]byte, length)
	copy(out, raw[dataStart:dataEnd])

	return out, nil
}

// decodeBytesArrayParam decodes a `bytes[]` ABI parameter whose head pointer
// resides at headOffset inside raw.  Returns the decoded slice of byte slices.
func decodeBytesArrayParam(raw []byte, headOffset int) ([][]byte, error) {
	// Read pointer to the start of the array encoding (relative to start of raw).
	arrayStart, err := readUint64Word(raw, headOffset)
	if err != nil {
		return nil, fmt.Errorf("abi: bytes[] pointer: %w", err)
	}

	// Read the number of elements.
	n, err := readUint64Word(raw, int(arrayStart))
	if err != nil {
		return nil, fmt.Errorf("abi: bytes[] length: %w", err)
	}

	result := make([][]byte, 0, n)

	// Read each element.  Element offsets are relative to arrayStart.
	for i := range n {
		elemPtrOffset := int(arrayStart) + wordSize + int(i)*wordSize

		elemRelOffset, err := readUint64Word(raw, elemPtrOffset)
		if err != nil {
			return nil, fmt.Errorf("abi: bytes[] element %d pointer: %w", i, err)
		}

		elemAbsOffset := int(arrayStart) + int(elemRelOffset)

		elemLen, err := readUint64Word(raw, elemAbsOffset)
		if err != nil {
			return nil, fmt.Errorf("abi: bytes[] element %d length: %w", i, err)
		}

		dataStart := elemAbsOffset + wordSize
		dataEnd := dataStart + int(elemLen)

		if len(raw) < dataEnd {
			return nil, fmt.Errorf("abi: bytes[] element %d data truncated", i)
		}

		elem := make([]byte, elemLen)
		copy(elem, raw[dataStart:dataEnd])
		result = append(result, elem)
	}

	return result, nil
}

// multiSendEntry represents a single packed transaction inside a multiSend call.
type multiSendEntry struct {
	// To is the 20-byte destination address for this sub-transaction.
	To [20]byte
	// Value is the ETH amount in wei for this sub-transaction.
	Value *big.Int
	// Data is the calldata for this sub-transaction.
	Data []byte
}

// decodeMultiSendTransactions decodes the packed `transactions` bytes produced
// by Gnosis Safe's multiSend / multiSendCallOnly encoding:
//
//	repeat { operation(1) | to(20) | value(32) | dataLength(32) | data(dataLength) }
func decodeMultiSendTransactions(transactions []byte) ([]multiSendEntry, error) {
	const headerSize = 1 + 20 + 32 + 32 //nolint:mnd  // operation + to + value + dataLength

	var entries []multiSendEntry

	pos := 0
	for pos < len(transactions) {
		if len(transactions)-pos < headerSize {
			return nil, fmt.Errorf("multisend: truncated entry header at pos %d", pos)
		}

		// Skip operation (1 byte).
		pos++

		var to [20]byte
		copy(to[:], transactions[pos:pos+20]) //nolint:mnd
		pos += 20                             //nolint:mnd

		value := new(big.Int).SetBytes(transactions[pos : pos+32]) //nolint:mnd
		pos += 32                                                  //nolint:mnd

		dataLen := int(new(big.Int).SetBytes(transactions[pos : pos+32]).Int64()) //nolint:mnd
		pos += 32                                                                 //nolint:mnd

		if len(transactions)-pos < dataLen {
			return nil, fmt.Errorf("multisend: truncated entry data at pos %d", pos)
		}

		data := make([]byte, dataLen)
		copy(data, transactions[pos:pos+dataLen])
		pos += dataLen

		entries = append(entries, multiSendEntry{To: to, Value: value, Data: data})
	}

	return entries, nil
}
