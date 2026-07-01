package common_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/LiquidCats/paw/services/watcher/internal/adapter/rpc/evm/data/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ====================
// Struct with embedded types for JSON tests
// ====================

type testStruct struct {
	BlockNumber *common.Big    `json:"blockNumber"`
	GasLimit    *common.Uint64 `json:"gasLimit"`
}

func TestEmbeddedTypes_JSONMarshalUnmarshal(t *testing.T) {
	blockNum := common.Big(*big.NewInt(12345678))
	gasLimit := common.Uint64(8000000)

	original := testStruct{
		BlockNumber: &blockNum,
		GasLimit:    &gasLimit,
	}

	marshaled, err := json.Marshal(original)
	require.NoError(t, err)

	var unmarshaled testStruct
	err = json.Unmarshal(marshaled, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, 0, original.BlockNumber.ToInt().Cmp(unmarshaled.BlockNumber.ToInt()))
	assert.Equal(t, *original.GasLimit, *unmarshaled.GasLimit)
}

func TestEmbeddedTypes_JSONFromRawString(t *testing.T) {
	jsonStr := `{"blockNumber":"0xbc614e","gasLimit":"0x7a1200"}`

	var s testStruct
	err := json.Unmarshal([]byte(jsonStr), &s)
	require.NoError(t, err)

	assert.Equal(t, int64(12345678), s.BlockNumber.ToInt().Int64())
	assert.Equal(t, common.Uint64(8000000), *s.GasLimit)
}
