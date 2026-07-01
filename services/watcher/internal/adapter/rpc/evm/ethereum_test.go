package evm_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LiquidCats/paw/services/watcher/configs"
	"github.com/LiquidCats/paw/services/watcher/internal/adapter/rpc/evm"
	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
	"github.com/LiquidCats/paw/services/watcher/test/fixtures"
	"github.com/LiquidCats/paw/services/watcher/test/mocks"
	"github.com/LiquidCats/paw/services/watcher/test/server"
	"github.com/rotisserie/eris"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	cfg := configs.RPCConfig{NodeURL: "http://example.invalid"}
	getter := mocks.NewMockGetter(t)

	c := evm.NewClient(cfg, "ETH", getter)
	require.NotNil(t, c)
}

func TestClient_GetMempool(t *testing.T) {
	t.Parallel()

	cfg := configs.RPCConfig{NodeURL: "http://example.invalid"}
	getter := mocks.NewMockGetter(t)

	c := evm.NewClient(cfg, "ETH", getter)

	got, err := c.GetMempool(context.Background())
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestClient_GetLatestBlock_Success(t *testing.T) {
	t.Parallel()

	srv := server.NewJSONRPCServer(t,
		server.RPCMocker{
			Method: "eth_getBlockByNumber",
			Result: fixtures.ETHBlockWithoutTransactions,
			ParamsCheck: func(t *testing.T, params []any) {
				require.Len(t, params, 2)
				require.Equal(t, "latest", params[0])
				require.Equal(t, false, params[1])
			},
		},
	)
	t.Cleanup(srv.Close)

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	getter := mocks.NewMockGetter(t)

	c := evm.NewClient(cfg, "ETH", getter)

	block, err := c.GetLatestBlock(context.Background())
	require.NoError(t, err)
	require.NotNil(t, block)

	assert.Equal(t, entities.BlockHash("0x9a1ae64d97306e7e0d266991f1de8f5250273dd65a2da84b8f4f68445662cfe3"), block.Hash)
	assert.Equal(t, entities.BlockHash("0xc31abe8647de68f9827c8294d2cc8b94723eeb277e796f65b6ec18a3453bfe8d"), block.PrevHash)
	assert.Equal(t, entities.BlockHeight(24_157_279), block.Height)
	assert.Equal(t, []entities.TxID{
		"0x80e12d2379c31433b944a6f9ebf66faa21a4dc53cda4721844e8adb4b9587c9c",
		"0xda8457ba09a7323a35cc59bb209dc3ccb57c1f454672f303cd30cb061fbfca73",
		"0x1126daf7e7385a2a35cc2dd2c80451005f516740e18911f8db5367a79f707b1c",
	}, block.Transactions)
}

func TestClient_GetLatestBlock_RPCError(t *testing.T) {
	t.Parallel()

	// Malformed response -> client should return an error
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":`)) // intentionally broken JSON
	}))
	t.Cleanup(srv.Close)

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	getter := mocks.NewMockGetter(t)

	c := evm.NewClient(cfg, "ETH", getter)

	_, err := c.GetLatestBlock(context.Background())
	require.Error(t, err)
}

func TestClient_GetBlockByHash_Success(t *testing.T) {
	t.Parallel()

	wantHash := entities.BlockHash("0x9a1ae64d97306e7e0d266991f1de8f5250273dd65a2da84b8f4f68445662cfe3")

	srv := server.NewJSONRPCServer(t,
		server.RPCMocker{
			Method: "eth_getBlockByHash",
			Result: fixtures.ETHBlockWithoutTransactions,
			ParamsCheck: func(t *testing.T, params []any) {
				require.Len(t, params, 2)
				require.Equal(t, string(wantHash), params[0])
				require.Equal(t, false, params[1])
			},
		},
	)
	t.Cleanup(srv.Close)

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	getter := mocks.NewMockGetter(t)

	c := evm.NewClient(cfg, "ETH", getter)

	block, err := c.GetBlockByHash(context.Background(), wantHash)
	require.NoError(t, err)
	require.NotNil(t, block)

	assert.Equal(t, entities.BlockHash("0x9a1ae64d97306e7e0d266991f1de8f5250273dd65a2da84b8f4f68445662cfe3"), block.Hash)
	assert.Equal(t, entities.BlockHash("0xc31abe8647de68f9827c8294d2cc8b94723eeb277e796f65b6ec18a3453bfe8d"), block.PrevHash)
	assert.Equal(t, entities.BlockHeight(24_157_279), block.Height)
	assert.Equal(t, []entities.TxID{
		"0x80e12d2379c31433b944a6f9ebf66faa21a4dc53cda4721844e8adb4b9587c9c",
		"0xda8457ba09a7323a35cc59bb209dc3ccb57c1f454672f303cd30cb061fbfca73",
		"0x1126daf7e7385a2a35cc2dd2c80451005f516740e18911f8db5367a79f707b1c",
	}, block.Transactions)
}

func TestClient_GetTransactionByTxID_USDTTransfer(t *testing.T) {
	t.Parallel()

	const (
		txHash   = entities.TxID("0xda8457ba09a7323a35cc59bb209dc3ccb57c1f454672f303cd30cb061fbfca73")
		fromAddr = entities.Address("0x814199Fa64E0bd5fDF2dfFF6e02489eAc0d26056")
		toAddr   = entities.Address("0x6Ab90575dCD3b90b916E2D005C2a7DCAd7E72C66")
	)

	contractsGetter := mocks.NewMockGetter(t)
	contractsGetter.EXPECT().
		GetContractInfoByAddress(mock.Anything, entities.Address("0xdAC17F958D2ee523a2206206994597C13D831ec7")).
		RunAndReturn(func(_ context.Context, addr entities.Address) (*entities.Contract, error) {
			// buildTransaction calls AddressToChecksumAddress on txLog.Address
			// and passes it to the getter.
			require.NotEmpty(t, addr)

			return &entities.Contract{
				Address:  "0xdAC17F958D2ee523a2206206994597C13D831ec7",
				Ticker:   "USDT",
				Decimals: 6,
			}, nil
		}).
		Once()

	srv := server.NewJSONRPCServer(t,
		server.RPCMocker{
			Method: "eth_getTransactionByHash",
			ParamsCheck: func(t *testing.T, params []any) {
				require.Len(t, params, 1)
				require.Equal(t, string(txHash), params[0])
			},
			Result: fixtures.ETHTransactionUSDT,
		},
		server.RPCMocker{
			Method: "eth_getTransactionReceipt",
			ParamsCheck: func(t *testing.T, params []any) {
				require.Len(t, params, 1)
				require.Equal(t, string(txHash), params[0])
			},
			Result: fixtures.ETHTransactionReceiptUSDT,
		},
	)
	t.Cleanup(srv.Close)

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := evm.NewClient(cfg, "ETH", contractsGetter)

	tx, err := c.GetTransactionByTxID(context.Background(), txHash)
	require.NoError(t, err)
	require.NotNil(t, tx)

	// Base ETH transfer always exists
	require.GreaterOrEqual(t, len(tx.Inputs), 1)
	require.Len(t, tx.Outputs, 2)

	assert.Equal(t, txHash, tx.TxID)

	assert.Equal(t, txHash, tx.TxID)
	assert.Equal(t, fromAddr, tx.Inputs[0].Address)
	assert.Equal(t, entities.Address("0xdAC17F958D2ee523a2206206994597C13D831ec7"), tx.Outputs[0].Address)
	assert.Equal(t, "0", tx.Outputs[0].Value.String()) // 0 ETH

	assert.Equal(t, fromAddr, tx.Inputs[1].Address)
	assert.Equal(t, toAddr, tx.Outputs[1].Address)
	assert.Equal(t, "27.85", tx.Outputs[1].Value.String())         // 27.85 USDT
	assert.Equal(t, entities.Ticker("USDT"), tx.Outputs[1].Ticker) // 27.85 USDT
}

func TestClient_GetTransactionByTxID_WhenContractLookupFails(t *testing.T) {
	t.Parallel()

	const (
		txHash   = entities.TxID("0xda8457ba09a7323a35cc59bb209dc3ccb57c1f454672f303cd30cb061fbfca73")
		fromAddr = entities.Address("0x814199Fa64E0bd5fDF2dfFF6e02489eAc0d26056")
	)

	contractsGetter := mocks.NewMockGetter(t)
	contractsGetter.EXPECT().
		GetContractInfoByAddress(mock.Anything, entities.Address("0xdAC17F958D2ee523a2206206994597C13D831ec7")).
		Return(nil, eris.New("not found")).
		Once()

	srv := server.NewJSONRPCServer(t,
		server.RPCMocker{
			Method: "eth_getTransactionByHash",
			ParamsCheck: func(t *testing.T, params []any) {
				require.Len(t, params, 1)
				require.Equal(t, string(txHash), params[0])
			},
			Result: fixtures.ETHTransactionUSDT,
		},
		server.RPCMocker{
			Method: "eth_getTransactionReceipt",
			ParamsCheck: func(t *testing.T, params []any) {
				require.Len(t, params, 1)
				require.Equal(t, string(txHash), params[0])
			},
			Result: fixtures.ETHTransactionReceiptUSDT,
		},
	)
	t.Cleanup(srv.Close)

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := evm.NewClient(cfg, "ETH", contractsGetter)

	tx, err := c.GetTransactionByTxID(context.Background(), txHash)
	require.Error(t, err)
	require.Nil(t, tx)
}

func TestClient_GetBlockByHashWithTransactions_ReorgMismatchCount(t *testing.T) {
	t.Parallel()

	const blockHash = entities.BlockHash("0x9a1ae64d97306e7e0d266991f1de8f5250273dd65a2da84b8f4f68445662cfe3")

	contractsGetter := mocks.NewMockGetter(t)

	// This method makes 2 RPC calls; the second one expects receipts count == tx count.
	srv := server.NewJSONRPCServer(t,
		server.RPCMocker{
			Method: "eth_getBlockByHash",
			ParamsCheck: func(t *testing.T, params []any) {
				require.Len(t, params, 2)
				require.Equal(t, string(blockHash), params[0])
				require.Equal(t, true, params[1])
			},
			Result: fixtures.ETHBlockWithTransactions,
		},
		server.RPCMocker{
			Method: "eth_getBlockReceipts",
			ParamsCheck: func(t *testing.T, params []any) {
				require.Len(t, params, 1)
				require.Equal(t, "0x1709c5f", params[0])
			},
			Result: `{"jsonrpc":"2.0","id":1,"result":[]}`,
		},
	)
	t.Cleanup(srv.Close)

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := evm.NewClient(cfg, "ETH", contractsGetter)

	_, err := c.GetBlockByHashWithTransactions(context.Background(), blockHash)
	require.Error(t, err)
}
