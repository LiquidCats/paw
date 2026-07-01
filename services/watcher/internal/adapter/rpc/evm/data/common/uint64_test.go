package common_test

import (
	"encoding/json"
	"testing"

	"github.com/LiquidCats/paw/watcher/internal/adapter/rpc/evm/data/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUint64_MarshalText(t *testing.T) {
	tests := []struct {
		name     string
		value    common.Uint64
		expected string
	}{
		{
			name:     "zero",
			value:    common.Uint64(0),
			expected: "0x0",
		},
		{
			name:     "small number",
			value:    common.Uint64(255),
			expected: "0xff",
		},
		{
			name:     "large number",
			value:    common.Uint64(1000000),
			expected: "0xf4240",
		},
		{
			name:     "max uint64",
			value:    common.Uint64(^uint64(0)),
			expected: "0xffffffffffffffff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.value.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestUint64_UnmarshalText(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    common.Uint64
		expectError bool
		errorType   error
	}{
		{
			name:     "valid hex",
			input:    "0xff",
			expected: common.Uint64(255),
		},
		{
			name:     "valid uppercase hex",
			input:    "0xFF",
			expected: common.Uint64(255),
		},
		{
			name:     "zero",
			input:    "0x0",
			expected: common.Uint64(0),
		},
		{
			name:     "max uint64",
			input:    "0xffffffffffffffff",
			expected: common.Uint64(^uint64(0)),
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
			name:        "too large for uint64",
			input:       "0x10000000000000000",
			expectError: true,
			errorType:   common.ErrUint64Range,
		},
		{
			name:     "empty string",
			input:    "",
			expected: common.Uint64(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var u common.Uint64
			err := u.UnmarshalText([]byte(tt.input))
			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != nil {
					require.ErrorIs(t, err, tt.errorType)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, u)
			}
		})
	}
}

func TestUint64_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    common.Uint64
		expectError bool
	}{
		{
			name:     "valid json string",
			input:    `"0xff"`,
			expected: common.Uint64(255),
		},
		{
			name:     "zero json string",
			input:    `"0x0"`,
			expected: common.Uint64(0),
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
			var u common.Uint64
			err := u.UnmarshalJSON([]byte(tt.input))
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, u)
			}
		})
	}
}

func TestUint64_String(t *testing.T) {
	tests := []struct {
		name     string
		value    common.Uint64
		expected string
	}{
		{
			name:     "zero",
			value:    common.Uint64(0),
			expected: "0x0",
		},
		{
			name:     "small number",
			value:    common.Uint64(16),
			expected: "0x10",
		},
		{
			name:     "large number",
			value:    common.Uint64(1000000),
			expected: "0xf4240",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.value.String())
		})
	}
}

func TestEncodeUint64(t *testing.T) {
	tests := []struct {
		name     string
		value    uint64
		expected string
	}{
		{
			name:     "zero",
			value:    0,
			expected: "0x0",
		},
		{
			name:     "small number",
			value:    16,
			expected: "0x10",
		},
		{
			name:     "max uint64",
			value:    ^uint64(0),
			expected: "0xffffffffffffffff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := common.EncodeUint64(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUint64_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		value common.Uint64
	}{
		{
			name:  "zero",
			value: common.Uint64(0),
		},
		{
			name:  "small number",
			value: common.Uint64(42),
		},
		{
			name:  "large number",
			value: common.Uint64(1000000000),
		},
		{
			name:  "max uint64",
			value: common.Uint64(^uint64(0)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := json.Marshal(&tt.value)
			require.NoError(t, err)

			var unmarshaled common.Uint64
			err = json.Unmarshal(marshaled, &unmarshaled)
			require.NoError(t, err)

			assert.Equal(t, tt.value, unmarshaled)
		})
	}
}

func TestUint64_UnmarshalText_AllNibbles(t *testing.T) {
	// Test all valid hex characters
	tests := []struct {
		input    string
		expected common.Uint64
	}{
		{input: "0x0", expected: 0},
		{input: "0x1", expected: 1},
		{input: "0x2", expected: 2},
		{input: "0x3", expected: 3},
		{input: "0x4", expected: 4},
		{input: "0x5", expected: 5},
		{input: "0x6", expected: 6},
		{input: "0x7", expected: 7},
		{input: "0x8", expected: 8},
		{input: "0x9", expected: 9},
		{input: "0xa", expected: 10},
		{input: "0xb", expected: 11},
		{input: "0xc", expected: 12},
		{input: "0xd", expected: 13},
		{input: "0xe", expected: 14},
		{input: "0xf", expected: 15},
		{input: "0xA", expected: 10},
		{input: "0xB", expected: 11},
		{input: "0xC", expected: 12},
		{input: "0xD", expected: 13},
		{input: "0xE", expected: 14},
		{input: "0xF", expected: 15},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var u common.Uint64
			err := u.UnmarshalText([]byte(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, u)
		})
	}
}

func TestUint64_UnmarshalText_InvalidNibbles(t *testing.T) {
	invalidInputs := []string{
		"0xg",
		"0xG",
		"0x!",
		"0x@",
		"0x-1",
		"0x+1",
		"0x ",
		"0x\t",
	}

	for _, input := range invalidInputs {
		t.Run(input, func(t *testing.T) {
			var u common.Uint64
			err := u.UnmarshalText([]byte(input))
			require.Error(t, err)
		})
	}
}
