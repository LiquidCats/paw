package common_test

import (
	"testing"

	"github.com/LiquidCats/paw/watcher/internal/adapter/rpc/evm/data/common"
	"github.com/go-playground/assert/v2"
)

func TestDecError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrEmptyString",
			err:      common.ErrEmptyString,
			expected: "empty hex string",
		},
		{
			name:     "ErrSyntax",
			err:      common.ErrSyntax,
			expected: "invalid hex string",
		},
		{
			name:     "ErrMissingPrefix",
			err:      common.ErrMissingPrefix,
			expected: "hex string without 0x prefix",
		},
		{
			name:     "ErrOddLength",
			err:      common.ErrOddLength,
			expected: "hex string of odd length",
		},
		{
			name:     "ErrEmptyNumber",
			err:      common.ErrEmptyNumber,
			expected: "hex string \"0x\"",
		},
		{
			name:     "ErrLeadingZero",
			err:      common.ErrLeadingZero,
			expected: "hex number with leading zero digits",
		},
		{
			name:     "ErrUint64Range",
			err:      common.ErrUint64Range,
			expected: "hex number > 64 bits",
		},
		{
			name:     "ErrBig256Range",
			err:      common.ErrBig256Range,
			expected: "hex number > 256 bits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestErrors_AreDistinct(t *testing.T) {
	errors := []error{
		common.ErrEmptyString,
		common.ErrSyntax,
		common.ErrMissingPrefix,
		common.ErrOddLength,
		common.ErrEmptyNumber,
		common.ErrLeadingZero,
		common.ErrUint64Range,
		common.ErrBig256Range,
	}

	for i, err1 := range errors {
		for j, err2 := range errors {
			if i != j {
				assert.NotEqual(t, err1.Error(), err2.Error())
			}
		}
	}
}
