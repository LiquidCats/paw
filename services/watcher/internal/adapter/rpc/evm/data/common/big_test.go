package common_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/LiquidCats/paw/watcher/internal/adapter/rpc/evm/data/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBig_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		value    *big.Int
		expected string
	}{
		{
			name:     "zero value",
			value:    big.NewInt(0),
			expected: "0x0",
		},
		{
			name:     "positive integer",
			value:    big.NewInt(255),
			expected: "0xff",
		},
		{
			name:     "large positive integer",
			value:    big.NewInt(1000000),
			expected: "0xf4240",
		},
		{
			name:     "negative integer",
			value:    big.NewInt(-255),
			expected: "-0xff",
		},
		{
			name:     "one",
			value:    big.NewInt(1),
			expected: "0x1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := common.Big(*tt.value)
			result, err := b.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestBig_UnmarshalText(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *big.Int
		expectError bool
		errorType   error
	}{
		{
			name:     "valid hex",
			input:    "0xff",
			expected: big.NewInt(255),
		},
		{
			name:     "valid uppercase hex",
			input:    "0xFF",
			expected: big.NewInt(255),
		},
		{
			name:     "zero",
			input:    "0x0",
			expected: big.NewInt(0),
		},
		{
			name:     "large number",
			input:    "0xf4240",
			expected: big.NewInt(1000000),
		},
		{
			name:        "missing prefix",
			input:       "ff",
			expectError: true,
			errorType:   common.ErrMissingPrefix,
		},
		{
			name:        "empty after prefix",
			input:       "0x",
			expectError: true,
			errorType:   common.ErrEmptyNumber,
		},
		{
			name:        "leading zero",
			input:       "0x0ff",
			expectError: true,
			errorType:   common.ErrLeadingZero,
		},
		{
			name:        "invalid character",
			input:       "0xgg",
			expectError: true,
			errorType:   common.ErrSyntax,
		},
		{
			name:     "empty string",
			input:    "",
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b common.Big
			err := b.UnmarshalText([]byte(tt.input))
			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != nil {
					require.ErrorIs(t, err, tt.errorType)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, 0, tt.expected.Cmp(b.ToInt()))
			}
		})
	}
}

func TestBig_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *big.Int
		expectError bool
	}{
		{
			name:     "valid json string",
			input:    `"0xff"`,
			expected: big.NewInt(255),
		},
		{
			name:     "zero json string",
			input:    `"0x0"`,
			expected: big.NewInt(0),
		},
		{
			name:        "non-string json",
			input:       `255`,
			expectError: true,
		},
		{
			name:        "invalid hex in json",
			input:       `"0xzz"`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b common.Big
			err := b.UnmarshalJSON([]byte(tt.input))
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, 0, tt.expected.Cmp(b.ToInt()))
			}
		})
	}
}

func TestBig_ToInt(t *testing.T) {
	expected := big.NewInt(12345)
	b := common.Big(*expected)

	result := b.ToInt()

	assert.Equal(t, 0, expected.Cmp(result))
}

func TestBig_String(t *testing.T) {
	tests := []struct {
		name     string
		value    *big.Int
		expected string
	}{
		{
			name:     "zero",
			value:    big.NewInt(0),
			expected: "0x0",
		},
		{
			name:     "positive",
			value:    big.NewInt(255),
			expected: "0xff",
		},
		{
			name:     "negative",
			value:    big.NewInt(-255),
			expected: "-0xff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := common.Big(*tt.value)
			assert.Equal(t, tt.expected, b.String())
		})
	}
}

func TestEncodeBig(t *testing.T) {
	tests := []struct {
		name     string
		value    *big.Int
		expected string
	}{
		{
			name:     "zero",
			value:    big.NewInt(0),
			expected: "0x0",
		},
		{
			name:     "positive number",
			value:    big.NewInt(16),
			expected: "0x10",
		},
		{
			name:     "negative number",
			value:    big.NewInt(-16),
			expected: "-0x10",
		},
		{
			name:     "large number",
			value:    new(big.Int).Exp(big.NewInt(2), big.NewInt(64), nil),
			expected: "0x10000000000000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.EncodeBig(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBig_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		value *big.Int
	}{
		{
			name:  "zero",
			value: big.NewInt(0),
		},
		{
			name:  "small positive",
			value: big.NewInt(42),
		},
		{
			name:  "large positive",
			value: new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := common.Big(*tt.value)

			marshaled, err := json.Marshal(&original)
			require.NoError(t, err)

			var unmarshaled common.Big
			err = json.Unmarshal(marshaled, &unmarshaled)
			require.NoError(t, err)

			assert.Equal(t, 0, original.ToInt().Cmp(unmarshaled.ToInt()))
		})
	}
}

func TestBig_UnmarshalText_TooLarge(t *testing.T) {
	// 257 bits (65 hex chars) - should fail
	input := "0x" + "f" + "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	var b common.Big
	err := b.UnmarshalText([]byte(input))
	assert.ErrorIs(t, err, common.ErrBig256Range)
}

func TestBig_UnmarshalText_MixedCase(t *testing.T) {
	tests := []struct {
		input    string
		expected *big.Int
	}{
		{input: "0xAbCdEf", expected: big.NewInt(0xABCDEF)},
		{input: "0xaBcDeF", expected: big.NewInt(0xABCDEF)},
		{input: "0XABCDEF", expected: big.NewInt(0xABCDEF)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var b common.Big
			err := b.UnmarshalText([]byte(tt.input))
			require.NoError(t, err)
			assert.Equal(t, 0, tt.expected.Cmp(b.ToInt()))
		})
	}
}

func TestBig_256BitBoundary(t *testing.T) {
	// Exactly 256 bits (64 hex chars) - should succeed
	max256 := "0x" + "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	var b common.Big
	err := b.UnmarshalText([]byte(max256))
	require.NoError(t, err)

	expected, _ := new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)
	assert.Equal(t, 0, expected.Cmp(b.ToInt()))
}
