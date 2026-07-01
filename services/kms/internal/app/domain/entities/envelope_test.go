package entities_test

//
// import (
//	"bytes"
//	"testing"
//
//	"github.com/awnumar/memguard"
//)
//
//// sealAndMarshal is a test helper that seals a small payload and returns the
//// encoded binary so individual bytes can be surgically corrupted.
// func sealAndMarshal(t *testing.T) []byte {
//	t.Helper()
//	sealing := newTestSealing(testParams)
//	pp := memguard.NewBufferFromBytes([]byte("pass"))
//	defer pp.Destroy()
//	data := memguard.NewBufferFromBytes([]byte("data"))
//	defer data.Destroy()
//
//	env, err := sealing.Seal(pp, data)
//	if err != nil {
//		t.Fatalf("Seal() error = %v", err)
//	}
//	defer env.Destroy()
//
//	b, err := env.MarshalBinary()
//	if err != nil {
//		t.Fatalf("MarshalBinary() error = %v", err)
//	}
//	return b
//}
//
// func TestEnvelopeMarshalBinarySize(t *testing.T) {
//	const payload = "hello"
//	sealing := newTestSealing(testParams)
//	pp := memguard.NewBufferFromBytes([]byte("pass"))
//	defer pp.Destroy()
//	data := memguard.NewBufferFromBytes([]byte(payload))
//	defer data.Destroy()
//
//	env, err := sealing.Seal(pp, data)
//	if err != nil {
//		t.Fatalf("Seal() error = %v", err)
//	}
//	defer env.Destroy()
//
//	b, err := env.MarshalBinary()
//	if err != nil {
//		t.Fatalf("MarshalBinary() error = %v", err)
//	}
//
//	want := argon2.HeaderSize + len(payload) + argon2.TagSize
//	if len(b) != want {
//		t.Fatalf("MarshalBinary() length = %d, want %d (HeaderSize=%d + payload=%d + tag=%d)",
//			len(b), want, argon2.HeaderSize, len(payload), argon2.TagSize)
//	}
//}
//
// func TestEnvelopeUnmarshalBinaryErrors(t *testing.T) {
//	valid := sealAndMarshal(t)
//
//	corruptAt := func(offset int) []byte {
//		b := bytes.Clone(valid)
//		b[offset] ^= 0xFF
//		return b
//	}
//	zeroRange := func(start, end int) []byte {
//		b := bytes.Clone(valid)
//		for i := start; i < end; i++ {
//			b[i] = 0
//		}
//		return b
//	}
//
//	tests := []struct {
//		name string
//		data []byte
//	}{
//		// Byte offsets in the on-disk format (see HeaderSize comment in sealing.go):
//		//   0-3   magic
//		//   4     version
//		//   5     kdf_id
//		//   6     aead_id
//		//   7-10  argon2 memory
//		//   11-14 argon2 iterations
//		//   15    argon2 parallelism
//		//   16-47 salt
//		{"truncated", valid[:argon2.HeaderSize+argon2.TagSize-1]},
//		{"bad magic", corruptAt(0)},
//		{"wrong version", corruptAt(4)},
//		{"wrong kdf", corruptAt(5)},
//		{"wrong aead", corruptAt(6)},
//		{"invalid kdf params", zeroRange(7, 16)}, // zeros MemoryKiB, Iterations, Parallelism
//	}
//
//	for _, tc := range tests {
//		t.Run(tc.name, func(t *testing.T) {
//			env := newExpectedEnvelope()
//			defer env.Destroy()
//
//			defer func() {
//				if r := recover(); r != nil {
//					t.Fatalf("UnmarshalBinary() panicked: %v", r)
//				}
//			}()
//
//			if err := env.UnmarshalBinary(tc.data); err == nil {
//				t.Fatalf("UnmarshalBinary() error = nil, want error for %q", tc.name)
//			}
//		})
//	}
//}
//
//// TestEnvelopeMarshalUnmarshalRoundTrip verifies that marshal → unmarshal →
//// marshal produces identical bytes, covering all header fields and buffer contents.
// func TestEnvelopeMarshalUnmarshalRoundTrip(t *testing.T) {
//	sealing := newTestSealing(testParams)
//	pp := memguard.NewBufferFromBytes([]byte("secret"))
//	defer pp.Destroy()
//	data := memguard.NewBufferFromBytes([]byte("round-trip payload"))
//	defer data.Destroy()
//
//	env, err := sealing.Seal(pp, data)
//	if err != nil {
//		t.Fatalf("Seal() error = %v", err)
//	}
//	defer env.Destroy()
//
//	first, err := env.MarshalBinary()
//	if err != nil {
//		t.Fatalf("first MarshalBinary() error = %v", err)
//	}
//
//	// Clone first before UnmarshalBinary: memguard.NewBufferFromBytes zeroes
//	// the source slice after copying, which would corrupt our reference copy.
//	decoded := newExpectedEnvelope()
//	defer decoded.Destroy()
//	if err = decoded.UnmarshalBinary(bytes.Clone(first)); err != nil {
//		t.Fatalf("UnmarshalBinary() error = %v", err)
//	}
//
//	second, err := decoded.MarshalBinary()
//	if err != nil {
//		t.Fatalf("second MarshalBinary() error = %v", err)
//	}
//
//	if !bytes.Equal(first, second) {
//		t.Fatal("marshal → unmarshal → marshal produced different bytes")
//	}
//}
//
//// TestEnvelopeUnmarshalBinaryPreservesParams checks that KDF params survive the
//// binary round-trip (they are written as big-endian uint32s and read back).
// func TestEnvelopeUnmarshalBinaryPreservesParams(t *testing.T) {
//	customParams := argon2.Params{
//		MemoryKiB:   argon2.ParamsMinMemoryKiB,
//		Iterations:  2,
//		Parallelism: 2,
//	}
//	sealing := newTestSealing(customParams)
//	pp := memguard.NewBufferFromBytes([]byte("pass"))
//	defer pp.Destroy()
//	data := memguard.NewBufferFromBytes([]byte("payload"))
//	defer data.Destroy()
//
//	env, err := sealing.Seal(pp, data)
//	if err != nil {
//		t.Fatalf("Seal() error = %v", err)
//	}
//	defer env.Destroy()
//
//	encoded, err := env.MarshalBinary()
//	if err != nil {
//		t.Fatalf("MarshalBinary() error = %v", err)
//	}
//
//	decoded := newExpectedEnvelope()
//	defer decoded.Destroy()
//	if err := decoded.UnmarshalBinary(encoded); err != nil {
//		t.Fatalf("UnmarshalBinary() error = %v", err)
//	}
//
//	if decoded.Params != customParams {
//		t.Fatalf("Params = %+v, want %+v", decoded.Params, customParams)
//	}
//}
//
// func TestEnvelopeDestroyNilSafe(t *testing.T) {
//	env := &argon2.Envelope{}
//
//	defer func() {
//		if r := recover(); r != nil {
//			t.Fatalf("Destroy() panicked with nil fields: %v", r)
//		}
//	}()
//
//	env.Destroy()
//}
