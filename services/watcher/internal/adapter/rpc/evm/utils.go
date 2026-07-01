package evm

import (
	"math/big"

	"github.com/LiquidCats/paw/services/watcher/internal/adapter/rpc/evm/data"
	"github.com/shopspring/decimal"
)

func toDecimal(v *big.Int, x int32) decimal.Decimal {
	return decimal.NewFromBigInt(v, 0).Shift(-x)
}

func toETH(v decimal.Decimal) decimal.Decimal {
	return v.Shift(-18) //nolint:mnd
}

func calculateFee(transaction *data.Transaction, receipt *data.TransactionReceipt) decimal.Decimal {
	typeInt := transaction.Type.ToInt()
	gasUsed := decimal.NewFromBigInt(receipt.GasUsed.ToInt(), 0)

	if data.TxTypeEIP1559.Cmp(typeInt) == 0 {
		effectiveGasPrice := decimal.NewFromBigInt(receipt.EffectiveGasPrice.ToInt(), 0)
		return toETH(effectiveGasPrice.Mul(gasUsed))
	}

	if data.TxTypeLegacy0.Cmp(typeInt) == 0 || data.TxTypeLegacy1.Cmp(typeInt) == 0 {
		gasPrice := decimal.NewFromBigInt(transaction.GasPrice.ToInt(), 0)
		return toETH(gasPrice.Mul(gasUsed))
	}

	return decimal.Zero
}
