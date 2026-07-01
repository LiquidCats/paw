package evm // nolint:testpackage

import (
	"math/big"
	"testing"

	data2 "github.com/LiquidCats/paw/watcher/internal/adapter/rpc/evm/data"
	"github.com/LiquidCats/paw/watcher/internal/adapter/rpc/evm/data/common"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestToDecimal(t *testing.T) {
	tests := []struct {
		name     string
		value    *big.Int
		shift    int32
		expected string
	}{
		{
			name:     "zero value",
			value:    big.NewInt(0),
			shift:    18,
			expected: "0",
		},
		{
			name:     "1 ETH in wei",
			value:    new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil),
			shift:    18,
			expected: "1",
		},
		{
			name:     "0.5 ETH in wei",
			value:    new(big.Int).Div(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil), big.NewInt(2)),
			shift:    18,
			expected: "0.5",
		},
		{
			name:     "small value with 6 decimals",
			value:    big.NewInt(1000000),
			shift:    6,
			expected: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toDecimal(tt.value, tt.shift)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestToETH(t *testing.T) {
	tests := []struct {
		name     string
		weiValue string
		expected string
	}{
		{
			name:     "zero wei",
			weiValue: "0",
			expected: "0",
		},
		{
			name:     "1 ETH in wei",
			weiValue: "1000000000000000000",
			expected: "1",
		},
		{
			name:     "small wei value",
			weiValue: "21000",
			expected: "0.000000000000021",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := decimal.RequireFromString(tt.weiValue)
			result := toETH(dec)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestCalculateFee(t *testing.T) {
	tests := []struct {
		name        string
		txType      int64
		gasPrice    int64 // for type 0, 1
		gasUsed     int64
		effectiveGP int64 // for type 2
		expected    string
	}{
		{
			name:        "type 0 legacy transaction",
			txType:      0,
			gasPrice:    20000000000, // 20 gwei
			gasUsed:     21000,
			effectiveGP: 0,
			expected:    "0.00042",
		},
		{
			name:        "type 1 access list transaction",
			txType:      1,
			gasPrice:    30000000000, // 30 gwei
			gasUsed:     50000,
			effectiveGP: 0,
			expected:    "0.0015",
		},
		{
			name:        "type 2 EIP-1559 transaction",
			txType:      2,
			gasPrice:    0,
			gasUsed:     21000,
			effectiveGP: 26672013, // ~26.67 gwei
			expected:    "0.000000560112273",
		},
		{
			name:        "unknown type returns zero",
			txType:      99,
			gasPrice:    20000000000,
			gasUsed:     21000,
			effectiveGP: 26672013,
			expected:    "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txType := common.Big(*big.NewInt(tt.txType))
			gasPrice := common.Big(*big.NewInt(tt.gasPrice))
			effectiveGasPrice := common.Big(*big.NewInt(tt.effectiveGP))

			gesUsed := common.Big(*big.NewInt(tt.gasUsed))
			tx := &data2.Transaction{
				Type:     &txType,
				GasPrice: &gasPrice,
			}
			receipt := &data2.TransactionReceipt{
				GasUsed:           &gesUsed,
				EffectiveGasPrice: &effectiveGasPrice,
			}

			result := calculateFee(tx, receipt)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}
