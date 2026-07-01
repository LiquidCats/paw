package unsafe

import "unsafe"

// StringToBytes converts a string to []byte (zero-copy).
//
// WARNING: The returned slice must NOT be modified.
// The original string's lifetime must exceed the returned slice's lifetime.
func StringToBytes(s string) []byte {
	// StringData returns a pointer to the underlying bytes of str.
	// Slice creates a slice from a pointer and a length.
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BytesToString converts a byte slice to a string without copying the underlying data.
// WARNING: Use cautiously with immutable data.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
