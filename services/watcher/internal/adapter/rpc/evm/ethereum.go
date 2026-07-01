package evm

import (
	"context"
	"math/big"

	"github.com/LiquidCats/paw/lib/jsonrpc"
	"github.com/LiquidCats/paw/services/watcher/configs"
	"github.com/LiquidCats/paw/services/watcher/internal/adapter/rpc/evm/data"
	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/errors"
	"github.com/LiquidCats/paw/services/watcher/internal/app/port/contracts"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

const (
	topicTransfer = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
)

type Client struct {
	cfg          configs.RPCConfig
	nativeTicker entities.Ticker
	contracts    contracts.Getter
}

func NewClient(cfg configs.RPCConfig, nativeTicker entities.Ticker, contracts contracts.Getter) *Client {
	return &Client{
		cfg:          cfg,
		contracts:    contracts,
		nativeTicker: nativeTicker,
	}
}

func (c *Client) GetMempool(_ context.Context) ([]entities.TxID, error) {
	return nil, nil
}

func (c *Client) GetLatestBlock(ctx context.Context) (*entities.Block, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("block_hash", "latest").Msg("GetLatestBlock")

	blockResult, err := jsonrpc.
		NewRequest[[]any, data.Block[entities.TxID]]("eth_getBlockByNumber", []any{"latest", false}).
		Prepare(c.cfg.NodeURL, jsonrpc.WithContext(ctx)).
		Execute(nil)
	if err != nil {
		return nil, eris.Wrapf(err, "execute get latest block")
	}

	return &entities.Block{
		BlockHeader: entities.BlockHeader{
			Hash:     blockResult.Hash,
			Height:   entities.BlockHeight(blockResult.Number.ToInt().Uint64()),
			PrevHash: blockResult.ParentHash,
		},
		Transactions: blockResult.Transactions,
	}, nil
}

func (c *Client) GetBlockByHash(ctx context.Context, hash entities.BlockHash) (*entities.Block, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Any("block_hash", hash).Msg("GetBlockByHash")

	blockResult, err := jsonrpc.
		NewRequest[[]any, data.Block[entities.TxID]]("eth_getBlockByHash", []any{string(hash), false}).
		Prepare(c.cfg.NodeURL, jsonrpc.WithContext(ctx)).
		Execute(nil)
	if err != nil {
		return nil, eris.Wrapf(err, "execute get latest block")
	}

	return &entities.Block{
		BlockHeader: entities.BlockHeader{
			Hash:     blockResult.Hash,
			Height:   entities.BlockHeight(blockResult.Number.ToInt().Uint64()),
			PrevHash: blockResult.ParentHash,
		},
		Transactions: blockResult.Transactions,
	}, nil
}

func (c *Client) GetBlockByHashWithTransactions(
	ctx context.Context,
	hash entities.BlockHash,
) (*entities.BlockWithTransactions[entities.TransactionAccountInput], error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Any("block_hash", hash).Msg("GetBlockByHashWithTransactions")

	blockResult, err := jsonrpc.
		NewRequest[[]any, data.Block[data.Transaction]]("eth_getBlockByHash", []any{string(hash), true}).
		Prepare(c.cfg.NodeURL, jsonrpc.WithContext(ctx)).
		Execute(nil)
	if err != nil {
		return nil, eris.Wrapf(err, "execute get latest block")
	}

	receiptsResult, err := jsonrpc.
		NewRequest[[]any, data.TransactionReceipts]("eth_getBlockReceipts", []any{blockResult.Number.String()}).
		Prepare(c.cfg.NodeURL, jsonrpc.WithContext(ctx)).
		Execute(nil)
	if err != nil {
		return nil, eris.Wrapf(err, "execute get latest block")
	}

	if len(blockResult.Transactions) != len(*receiptsResult) {
		return nil, errors.NewPossibleReorgError(hash)
	}

	txMap := make(map[entities.TxID]*data.Transaction)

	for _, transaction := range blockResult.Transactions {
		txMap[transaction.Hash] = &transaction
	}

	blockEntity := entities.BlockWithTransactions[entities.TransactionAccountInput]{
		BlockHeader: entities.BlockHeader{
			Hash:     blockResult.Hash,
			Height:   entities.BlockHeight(blockResult.Number.ToInt().Uint64()),
			PrevHash: blockResult.ParentHash,
		},
		Transactions: make([]entities.Transaction[entities.TransactionAccountInput], len(blockResult.Transactions)),
	}

	for _, receipt := range *receiptsResult {
		transaction, ok := txMap[receipt.TransactionHash]
		if !ok {
			return nil, errors.NewPossibleReorgError(hash)
		}

		blockTransaction, err := c.buildTransaction(ctx, transaction, &receipt)
		if err != nil {
			return nil, eris.Wrapf(err, "build transaction")
		}

		blockEntity.Transactions = append(blockEntity.Transactions, *blockTransaction)
	}

	return &blockEntity, nil
}

func (c *Client) GetTransactionByTxID(
	ctx context.Context,
	hash entities.TxID,
) (*entities.Transaction[entities.TransactionAccountInput], error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Any("txid", hash).Msg("GetTransactionByTxID")

	transactionResult, err := jsonrpc.
		NewRequest[any, data.Transaction]("eth_getTransactionByHash", []any{hash}).
		Prepare(c.cfg.NodeURL, jsonrpc.WithContext(ctx)).
		Execute(nil)
	if err != nil {
		return nil, eris.Wrapf(err, "get block by hash %s", hash)
	}

	receiptResult, err := jsonrpc.
		NewRequest[any, data.TransactionReceipt]("eth_getTransactionReceipt", []any{hash}).
		Prepare(c.cfg.NodeURL, jsonrpc.WithContext(ctx)).
		Execute(nil)
	if err != nil {
		return nil, eris.Wrapf(err, "prepare get block by hash %s", hash)
	}

	transaction, err := c.buildTransaction(ctx, transactionResult, receiptResult)
	if err != nil {
		return nil, eris.Wrapf(err, "build transaction")
	}

	return transaction, nil
}

func (c *Client) buildTransaction(
	ctx context.Context,
	transaction *data.Transaction,
	receipt *data.TransactionReceipt,
) (*entities.Transaction[entities.TransactionAccountInput], error) {
	logger := zerolog.Ctx(ctx).With().Any("txid", transaction.Hash).Logger()

	blockTransaction := entities.Transaction[entities.TransactionAccountInput]{
		TxID:      transaction.Hash,
		BlockHash: transaction.BlockHash,
		Inputs:    []entities.TransactionAccountInput{},
		Outputs:   []entities.TransactionOutput{},
		Fee:       calculateFee(transaction, receipt),
	}

	input := entities.TransactionAccountInput{
		Address: data.AddressToChecksumAddress(transaction.From),
	}
	blockTransaction.Inputs = append(blockTransaction.Inputs, input)

	output := entities.TransactionOutput{
		N:       0,
		Value:   toDecimal(transaction.Value.ToInt(), 18), //nolint:mnd
		Ticker:  c.nativeTicker,
		Address: data.AddressToChecksumAddress(transaction.To),
	}
	blockTransaction.Outputs = append(blockTransaction.Outputs, output)

	for idx, txLog := range receipt.Logs {
		if txLog.Topics[0] != topicTransfer {
			continue
		}

		addressFrom := data.AddressToChecksumAddress(txLog.Topics[1])
		addressTo := data.AddressToChecksumAddress(txLog.Topics[2])
		contractAddress := data.AddressToChecksumAddress(txLog.Address)

		contractInfo, err := c.contracts.GetContractInfoByAddress(ctx, contractAddress)
		if err != nil {
			logger.Error().
				Any("error", eris.ToJSON(err, true)).
				Any("contract_address", contractAddress).
				Any("address_from", addressFrom).
				Any("address_to", addressTo).
				Msg("get contract info")
			return nil, eris.Wrap(err, "get contract info")
		}

		if contractInfo == nil {
			continue
		}

		if contractInfo.Ticker == "" {
			continue
		}

		logData, _ := new(big.Int).SetString(txLog.Data, 0) //nolint:mnd

		value := toDecimal(logData, contractInfo.Decimals)

		input := entities.TransactionAccountInput{
			Address: addressFrom,
		}
		blockTransaction.Inputs = append(blockTransaction.Inputs, input)

		output := entities.TransactionOutput{
			Value:    value,
			Address:  addressTo,
			Contract: contractInfo.Address,
			N:        uint32(idx), // nolint:gosec
			Ticker:   contractInfo.Ticker,
		}
		blockTransaction.Outputs = append(blockTransaction.Outputs, output)
	}
	return &blockTransaction, nil
}
