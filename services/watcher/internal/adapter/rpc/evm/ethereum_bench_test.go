package evm_test

import (
	"context"
	"testing"

	"github.com/LiquidCats/paw/watcher/configs"
	"github.com/LiquidCats/paw/watcher/internal/adapter/rpc/evm"
	"github.com/LiquidCats/paw/watcher/internal/app/domain/entities"
	"github.com/LiquidCats/paw/watcher/test/fixtures"
	"github.com/LiquidCats/paw/watcher/test/mocks"
	"github.com/LiquidCats/paw/watcher/test/server"
	"github.com/stretchr/testify/mock"
)

func BenchmarkClient_GetMempool(b *testing.B) {
	cfg := configs.RPCConfig{NodeURL: "http://example.invalid"}
	getter := mocks.NewMockGetter(&testing.T{})

	c := evm.NewClient(cfg, "ETH", getter)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_, _ = c.GetMempool(ctx)
	}
}

func BenchmarkClient_GetLatestBlock(b *testing.B) {
	srv := server.NewJSONRPCServer(&testing.T{},
		server.RPCMocker{
			Method: "eth_getBlockByNumber",
			Result: fixtures.ETHBlockWithoutTransactions,
		},
	)
	defer srv.Close()

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	getter := mocks.NewMockGetter(&testing.T{})

	c := evm.NewClient(cfg, "ETH", getter)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_, _ = c.GetLatestBlock(ctx)
	}
}

func BenchmarkClient_GetBlockByHash(b *testing.B) {
	srv := server.NewJSONRPCServer(&testing.T{},
		server.RPCMocker{
			Method: "eth_getBlockByHash",
			Result: fixtures.ETHBlockWithoutTransactions,
		},
	)
	defer srv.Close()

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	getter := mocks.NewMockGetter(&testing.T{})

	c := evm.NewClient(cfg, "ETH", getter)
	ctx := context.Background()
	hash := entities.BlockHash("0x9a1ae64d97306e7e0d266991f1de8f5250273dd65a2da84b8f4f68445662cfe3")

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_, _ = c.GetBlockByHash(ctx, hash)
	}
}

func BenchmarkClient_GetTransactionByTxID(b *testing.B) {
	contractsGetter := mocks.NewMockGetter(&testing.T{})
	contractsGetter.EXPECT().
		GetContractInfoByAddress(mock.Anything, entities.Address("0xdAC17F958D2ee523a2206206994597C13D831ec7")).
		Return(&entities.Contract{
			Address:  "0xdAC17F958D2ee523a2206206994597C13D831ec7",
			Ticker:   "USDT",
			Decimals: 6,
		}, nil).
		Maybe()

	srv := server.NewJSONRPCServer(&testing.T{},
		server.RPCMocker{
			Method: "eth_getTransactionByHash",
			Result: fixtures.ETHTransactionUSDT,
		},
		server.RPCMocker{
			Method: "eth_getTransactionReceipt",
			Result: fixtures.ETHTransactionReceiptUSDT,
		},
	)
	defer srv.Close()

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := evm.NewClient(cfg, "ETH", contractsGetter)
	ctx := context.Background()
	txID := entities.TxID("0xda8457ba09a7323a35cc59bb209dc3ccb57c1f454672f303cd30cb061fbfca73")

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_, _ = c.GetTransactionByTxID(ctx, txID)
	}
}

func BenchmarkClient_GetTransactionByTxID_Native(b *testing.B) {
	contractsGetter := mocks.NewMockGetter(&testing.T{})

	srv := server.NewJSONRPCServer(&testing.T{},
		server.RPCMocker{
			Method: "eth_getTransactionByHash",
			Result: fixtures.ETHTransactionNative,
		},
		server.RPCMocker{
			Method: "eth_getTransactionReceipt",
			Result: fixtures.ETHTransactionReceiptNative,
		},
	)
	defer srv.Close()

	cfg := configs.RPCConfig{NodeURL: srv.URL}
	c := evm.NewClient(cfg, "ETH", contractsGetter)
	ctx := context.Background()
	txID := entities.TxID("0x80e12d2379c31433b944a6f9ebf66faa21a4dc53cda4721844e8adb4b9587c9c")

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_, _ = c.GetTransactionByTxID(ctx, txID)
	}
}
