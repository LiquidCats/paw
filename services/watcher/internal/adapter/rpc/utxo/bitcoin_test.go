package utxo_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LiquidCats/paw/watcher/configs"
	"github.com/LiquidCats/paw/watcher/internal/adapter/rpc/utxo"
	"github.com/LiquidCats/paw/watcher/internal/app/domain/entities"
	"github.com/LiquidCats/paw/watcher/test/fixtures"
	"github.com/LiquidCats/paw/watcher/test/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTicker = entities.Ticker("BTC")

func TestNewClient(t *testing.T) {
	t.Parallel()

	cfg := configs.RPCConfig{NodeURL: "http://example.invalid"}

	c := utxo.NewClient(cfg, testTicker)
	require.NotNil(t, c)
}

func TestClient_GetMempool_Success(t *testing.T) {
	t.Parallel()

	srv := server.NewJSONRPCServer(t,
		server.RPCMocker{
			Method: "getrawmempool",
			Result: fixtures.BTCMempool,
			ParamsCheck: func(t *testing.T, params []any) {
				require.Len(t, params, 2)
				require.Equal(t, "latest", params[0])
				require.Equal(t, false, params[1])
			},
		},
	)
	t.Cleanup(srv.Close)

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := utxo.NewClient(cfg, testTicker)

	mempool, err := c.GetMempool(context.Background())
	require.NoError(t, err)
	require.NotNil(t, mempool)
	assert.Len(t, mempool, 13)
	assert.Equal(t, entities.TxID("abbe34887904e40479c9350c54194239cfe22b5eb4abdbb2b911bc1cd39c9276"), mempool[0])
}

func TestClient_GetMempool_RPCError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":`)) // intentionally broken JSON
	}))
	t.Cleanup(srv.Close)

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := utxo.NewClient(cfg, testTicker)

	_, err := c.GetMempool(context.Background())
	require.Error(t, err)
}

func TestClient_GetLatestBlock_Success(t *testing.T) {
	t.Parallel()

	srv := server.NewJSONRPCServer(t,
		server.RPCMocker{
			Method: "getbestblockhash",
			Result: `{"jsonrpc":"2.0","id":"1","result":"` + fixtures.BTCBlockHash + `"}`,
			ParamsCheck: func(t *testing.T, params []any) {
				require.Len(t, params, 2)
				require.Equal(t, "latest", params[0])
				require.Equal(t, false, params[1])
			},
		},
		server.RPCMocker{
			Method: "getblock",
			Result: fixtures.BTCBlockWithoutTransactions,
			ParamsCheck: func(t *testing.T, params []any) {
				require.Len(t, params, 2)
				require.Equal(t, fixtures.BTCBlockHash, params[0])
				require.Equal(t, float64(1), params[1]) // nolint:testifylint
			},
		},
	)
	t.Cleanup(srv.Close)

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := utxo.NewClient(cfg, testTicker)

	block, err := c.GetLatestBlock(context.Background())
	require.NoError(t, err)
	require.NotNil(t, block)

	assert.Equal(t, entities.BlockHash(fixtures.BTCBlockHash), block.Hash)
	assert.Equal(t, entities.BlockHeight(fixtures.BTCBlockHeight), block.Height)
	assert.Equal(t, entities.BlockHash("00000000000000000000fe63a464319078b834b64a3748c137985324b500470d"), block.PrevHash)
	assert.Len(t, block.Transactions, 5)
}

func TestClient_GetLatestBlock_GetBestBlockHashError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":`)) // broken JSON
	}))
	t.Cleanup(srv.Close)

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := utxo.NewClient(cfg, testTicker)

	_, err := c.GetLatestBlock(context.Background())
	require.Error(t, err)
}

func TestClient_GetBlockByHash_Success(t *testing.T) {
	t.Parallel()

	wantHash := entities.BlockHash(fixtures.BTCBlockHash)

	srv := server.NewJSONRPCServer(t,
		server.RPCMocker{
			Method: "getblock",
			Result: fixtures.BTCBlockWithoutTransactions,
			ParamsCheck: func(t *testing.T, params []any) {
				require.Len(t, params, 2)
				require.Equal(t, string(wantHash), params[0])
				require.Equal(t, float64(1), params[1]) // nolint:testifylint
			},
		},
	)
	t.Cleanup(srv.Close)

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := utxo.NewClient(cfg, testTicker)

	block, err := c.GetBlockByHash(context.Background(), wantHash)
	require.NoError(t, err)
	require.NotNil(t, block)

	assert.Equal(t, entities.BlockHash(fixtures.BTCBlockHash), block.Hash)
	assert.Equal(t, entities.BlockHeight(fixtures.BTCBlockHeight), block.Height)
	assert.Equal(t, entities.BlockHash("00000000000000000000fe63a464319078b834b64a3748c137985324b500470d"), block.PrevHash)
	assert.Len(t, block.Transactions, 5)
	assert.Equal(t, entities.TxID("50e2a04b362278269fd7e7ee56febf2163dca2cb03a46982a3bdcfb16b27b151"), block.Transactions[0])
}

func TestClient_GetBlockByHash_RPCError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":`)) // broken JSON
	}))
	t.Cleanup(srv.Close)

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := utxo.NewClient(cfg, testTicker)

	_, err := c.GetBlockByHash(context.Background(), "somehash")
	require.Error(t, err)
}

func TestClient_GetBlockByHashWithTransactions_Success(t *testing.T) {
	t.Parallel()

	wantHash := entities.BlockHash(fixtures.BTCBlockHash)

	srv := server.NewJSONRPCServer(t,
		server.RPCMocker{
			Method: "getblock",
			Result: fixtures.BTCBlockWithTransactions,
			ParamsCheck: func(t *testing.T, params []any) {
				require.Len(t, params, 2)
				require.Equal(t, string(wantHash), params[0])
				require.Equal(t, float64(2), params[1]) // nolint:testifylint
			},
		},
	)
	t.Cleanup(srv.Close)

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := utxo.NewClient(cfg, testTicker)

	block, err := c.GetBlockByHashWithTransactions(context.Background(), wantHash)
	require.NoError(t, err)
	require.NotNil(t, block)

	assert.Equal(t, entities.BlockHash(fixtures.BTCBlockHash), block.Hash)
	assert.Equal(t, entities.BlockHeight(fixtures.BTCBlockHeight), block.Height)
	assert.Equal(t, entities.BlockHash("00000000000000000000fe63a464319078b834b64a3748c137985324b500470d"), block.PrevHash)
	assert.Len(t, block.Transactions, 5)

	// Verify first transaction (coinbase)
	coinbaseTx := block.Transactions[0]
	assert.Equal(t, entities.TxID("50e2a04b362278269fd7e7ee56febf2163dca2cb03a46982a3bdcfb16b27b151"), coinbaseTx.TxID)
	assert.Len(t, coinbaseTx.Inputs, 1)
	assert.Len(t, coinbaseTx.Outputs, 4)
	assert.Equal(t, testTicker, coinbaseTx.Outputs[0].Ticker)
	assert.Equal(t, entities.Address("bc1qwzrryqr3ja8w7hnja2spmkgfdcgvqwp5swz4af4ngsjecfz0w0pqud7k38"), coinbaseTx.Outputs[0].Address)

	// Verify second transaction
	tx2 := block.Transactions[1]
	assert.Equal(t, entities.TxID("42817bca63a2e1594f62db5306323159c6804f95a585981e07b013d77c4139dc"), tx2.TxID)
	assert.Len(t, tx2.Inputs, 1)
	assert.Equal(t, entities.TxID("7de181eb3e16e9086dcf349e5397066bd5300dbeedb512869b04ffec6be6bee6"), tx2.Inputs[0].TxID)
	assert.Equal(t, uint32(2), tx2.Inputs[0].N)
	assert.Len(t, tx2.Outputs, 1)
	assert.Equal(t, entities.Address("1MNw8HGYgFT7fsSVPxPSu2Mj7ruQJ82Ls4"), tx2.Outputs[0].Address)
	assert.Equal(t, "0.0080698", tx2.Outputs[0].Value.String())
	assert.Equal(t, "0.0002", tx2.Fee.String())
}

func TestClient_GetBlockByHashWithTransactions_RPCError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":`)) // broken JSON
	}))
	t.Cleanup(srv.Close)

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := utxo.NewClient(cfg, testTicker)

	_, err := c.GetBlockByHashWithTransactions(context.Background(), "somehash")
	require.Error(t, err)
}

func TestClient_GetTransactionByTxID_Success(t *testing.T) {
	t.Parallel()

	wantTxID := entities.TxID(fixtures.BTCTxID)

	srv := server.NewJSONRPCServer(t,
		server.RPCMocker{
			Method: "getrawtransaction",
			Result: fixtures.BTCTransaction,
			ParamsCheck: func(t *testing.T, params []any) {
				require.Len(t, params, 2)
				require.Equal(t, string(wantTxID), params[0])
				require.Equal(t, float64(2), params[1]) // nolint:testifylint
			},
		},
	)
	t.Cleanup(srv.Close)

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := utxo.NewClient(cfg, testTicker)

	tx, err := c.GetTransactionByTxID(context.Background(), wantTxID)
	require.NoError(t, err)
	require.NotNil(t, tx)

	assert.Equal(t, wantTxID, tx.TxID)
	assert.Equal(t, entities.BlockHash(fixtures.BTCBlockHash), tx.BlockHash)
	assert.Len(t, tx.Inputs, 17)
	assert.Len(t, tx.Outputs, 1)

	// Verify first input
	assert.Equal(t, entities.TxID("bb7436aabaa11c0d543aaa5536e752b61dd6644ec0bd5d9b3c2ff61b06660410"), tx.Inputs[0].TxID)
	assert.Equal(t, uint32(0), tx.Inputs[0].N)

	// Verify output
	assert.Equal(t, entities.Address("bc1q306ym2z994vqnv3kyua0yxlq3gzmqeslmnjsh3"), tx.Outputs[0].Address)
	assert.Equal(t, "0.00634194", tx.Outputs[0].Value.String())
	assert.Equal(t, testTicker, tx.Outputs[0].Ticker)
}

func TestClient_GetTransactionByTxID_RPCError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":`)) // broken JSON
	}))
	t.Cleanup(srv.Close)

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := utxo.NewClient(cfg, testTicker)

	_, err := c.GetTransactionByTxID(context.Background(), "somehash")
	require.Error(t, err)
}
