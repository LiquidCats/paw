package configuration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LiquidCats/paw/services/litehsm/pkg/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLockedParam_UnmarshalText_Empty(t *testing.T) {
	var p configuration.LockedParam
	err := p.UnmarshalText([]byte{})
	require.NoError(t, err)
	assert.Nil(t, p.LockedBuffer)
}

func TestLockedParam_UnmarshalText_InvalidFormat(t *testing.T) {
	var p configuration.LockedParam
	err := p.UnmarshalText([]byte("file/tmp/secret"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sensitive param")
	assert.Contains(t, err.Error(), "colon")
}

func TestLockedParam_UnmarshalText_UnsupportedType(t *testing.T) {
	var p configuration.LockedParam
	err := p.UnmarshalText([]byte("XXXXX:value"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sensitive param")
	assert.Contains(t, err.Error(), "unsupported")
}

func TestLockedParam_UnmarshalText_ShortInput_Errors(t *testing.T) {
	var p configuration.LockedParam
	err := p.UnmarshalText([]byte("short"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "colon")
	assert.Nil(t, p.LockedBuffer)
}

func TestLockedParam_UnmarshalText_Env(t *testing.T) {
	t.Setenv("LITEHSM_TEST_SECRET", "secret")

	var p configuration.LockedParam
	err := p.UnmarshalText([]byte("envs:LITEHSM_TEST_SECRET"))
	require.NoError(t, err)
	require.NotNil(t, p.LockedBuffer)
	defer p.Destroy()

	assert.Equal(t, []byte("secret"), p.Bytes())
}

func TestLockedParam_UnmarshalText_File(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret.txt")
	require.NoError(t, os.WriteFile(path, []byte("secret"), 0o600))

	var p configuration.LockedParam
	err := p.UnmarshalText([]byte("file:" + path))
	require.NoError(t, err)
	require.NotNil(t, p.LockedBuffer)
	defer p.Destroy()

	assert.Equal(t, []byte("secret"), p.Bytes())
}

// ---------------------------------------------------------------------------
// SealedParam
// ---------------------------------------------------------------------------

func TestSealedParam_UnmarshalText_InvalidFormat(t *testing.T) {
	var p configuration.SealedParam
	err := p.UnmarshalText([]byte("file/tmp/secret"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sensitive param")
}

func TestSealedParam_UnmarshalText_UnsupportedType(t *testing.T) {
	var p configuration.SealedParam
	err := p.UnmarshalText([]byte("XXXXX:value"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestSealedParam_UnmarshalText_ShortInput_Errors(t *testing.T) {
	var p configuration.SealedParam
	err := p.UnmarshalText([]byte("short"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "colon")
	assert.Nil(t, p.Enclave)
}

func BenchmarkLockedParam_UnmarshalText_Empty(b *testing.B) {
	for b.Loop() {
		var p configuration.LockedParam
		_ = p.UnmarshalText(make([]byte, 0))
	}
}

func BenchmarkLockedParam_UnmarshalText_ColonError(b *testing.B) {
	src := []byte("file:/tmp/secret.txt")
	for b.Loop() {
		var p configuration.LockedParam
		input := make([]byte, len(src))
		copy(input, src)
		_ = p.UnmarshalText(input)
	}
}
