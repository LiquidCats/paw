package utxo_test

import (
	"context"
	"testing"

	"github.com/LiquidCats/paw/services/watcher/configs"
	"github.com/LiquidCats/paw/services/watcher/internal/adapter/rpc/utxo"
	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
	"github.com/LiquidCats/paw/services/watcher/test/fixtures"
	"github.com/LiquidCats/paw/services/watcher/test/server"
)

func BenchmarkClient_GetMempool(b *testing.B) {
	srv := server.NewJSONRPCServer(&testing.T{},
		server.RPCMocker{
			Method: "getrawmempool",
			Result: fixtures.BTCMempool,
		},
	)
	defer srv.Close()

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := utxo.NewClient(cfg, testTicker)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_, _ = c.GetMempool(ctx)
	}
}

func BenchmarkClient_GetBlockByHash(b *testing.B) {
	srv := server.NewJSONRPCServer(&testing.T{},
		server.RPCMocker{
			Method: "getblock",
			Result: fixtures.BTCBlockWithoutTransactions,
		},
	)
	defer srv.Close()

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := utxo.NewClient(cfg, testTicker)
	ctx := context.Background()
	hash := entities.BlockHash(fixtures.BTCBlockHash)

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_, _ = c.GetBlockByHash(ctx, hash)
	}
}

func BenchmarkClient_GetBlockByHashWithTransactions(b *testing.B) {
	srv := server.NewJSONRPCServer(&testing.T{},
		server.RPCMocker{
			Method: "getblock",
			Result: fixtures.BTCBlockWithTransactions,
		},
	)
	defer srv.Close()

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := utxo.NewClient(cfg, testTicker)
	ctx := context.Background()
	hash := entities.BlockHash(fixtures.BTCBlockHash)

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_, _ = c.GetBlockByHashWithTransactions(ctx, hash)
	}
}

func BenchmarkClient_GetTransactionByTxID(b *testing.B) {
	srv := server.NewJSONRPCServer(&testing.T{},
		server.RPCMocker{
			Method: "getrawtransaction",
			Result: fixtures.BTCTransaction,
		},
	)
	defer srv.Close()

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := utxo.NewClient(cfg, testTicker)
	ctx := context.Background()
	txID := entities.TxID(fixtures.BTCTxID)

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_, _ = c.GetTransactionByTxID(ctx, txID)
	}
}

func BenchmarkClient_GetLatestBlock(b *testing.B) {
	srv := server.NewJSONRPCServer(&testing.T{},
		server.RPCMocker{
			Method: "getbestblockhash",
			Result: `{"jsonrpc":"2.0","id":"1","result":"` + fixtures.BTCBlockHash + `"}`,
		},
		server.RPCMocker{
			Method: "getblock",
			Result: fixtures.BTCBlockWithoutTransactions,
		},
	)
	defer srv.Close()

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := utxo.NewClient(cfg, testTicker)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_, _ = c.GetLatestBlock(ctx)
	}
}
