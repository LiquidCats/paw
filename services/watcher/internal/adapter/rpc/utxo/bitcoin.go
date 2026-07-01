package utxo

import (
	"context"

	"github.com/LiquidCats/paw/lib/jsonrpc"
	"github.com/LiquidCats/paw/services/watcher/configs"
	"github.com/LiquidCats/paw/services/watcher/internal/adapter/rpc/utxo/data"
	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
	"github.com/rotisserie/eris"
)

type Client struct {
	nativeTicker entities.Ticker
	cfg          configs.RPCConfig
}

func NewClient(cfg configs.RPCConfig, nativeTicker entities.Ticker) *Client {
	return &Client{
		nativeTicker: nativeTicker,
		cfg:          cfg,
	}
}

func (c *Client) GetMempool(ctx context.Context) ([]entities.TxID, error) {
	mempoolResult, err := jsonrpc.
		NewRequest[[]any, []entities.TxID]("getrawmempool", []any{"latest", false}).
		Prepare(c.cfg.NodeURL, jsonrpc.WithContext(ctx)).
		Execute(nil)
	if err != nil {
		return nil, eris.Wrap(err, "GetMempool: execute")
	}

	return *mempoolResult, nil
}

func (c *Client) GetLatestBlock(ctx context.Context) (*entities.Block, error) {
	bestBlockHashResult, err := jsonrpc.
		NewRequest[[]any, entities.BlockHash]("getbestblockhash", []any{"latest", false}).
		Prepare(c.cfg.NodeURL, jsonrpc.WithContext(ctx)).
		Execute(nil)
	if err != nil {
		return nil, eris.Wrap(err, "GetLatestBlockHash: execute")
	}

	block, err := c.GetBlockByHash(ctx, *bestBlockHashResult)
	if err != nil {
		return nil, eris.Wrap(err, "GetLatestBlock: execute")
	}

	return block, nil
}

func (c *Client) GetBlockByHash(ctx context.Context, hash entities.BlockHash) (*entities.Block, error) {
	blockResult, err := jsonrpc.
		NewRequest[[]any, data.Block[entities.TxID]]("getblock", []any{hash, 1}).
		Prepare(c.cfg.NodeURL, jsonrpc.WithContext(ctx)).
		Execute(nil)
	if err != nil {
		return nil, eris.Wrap(err, "GetBlockByHash: execute")
	}

	block := entities.Block{
		BlockHeader: entities.BlockHeader{
			Height:   blockResult.Height,
			Hash:     blockResult.Hash,
			PrevHash: blockResult.PreviousBlockHash,
		},
		Transactions: blockResult.Tx,
	}

	return &block, nil
}

func (c *Client) GetBlockByHashWithTransactions(
	ctx context.Context,
	hash entities.BlockHash,
) (*entities.BlockWithTransactions[entities.TransactionUtxoInput], error) {
	blockResult, err := jsonrpc.
		NewRequest[[]any, data.Block[data.Transaction]]("getblock", []any{hash, 2}).
		Prepare(c.cfg.NodeURL, jsonrpc.WithContext(ctx)).
		Execute(nil)
	if err != nil {
		return nil, eris.Wrap(err, "GetBlockByHash: execute")
	}

	block := entities.BlockWithTransactions[entities.TransactionUtxoInput]{
		BlockHeader: entities.BlockHeader{
			Height:   blockResult.Height,
			Hash:     blockResult.Hash,
			PrevHash: blockResult.PreviousBlockHash,
		},
		Transactions: make([]entities.Transaction[entities.TransactionUtxoInput], 0, len(blockResult.Tx)),
	}

	for _, tx := range blockResult.Tx {
		transaction := c.buildTransaction(&tx)
		block.Transactions = append(block.Transactions, transaction)
	}

	return &block, nil
}

func (c *Client) GetTransactionByTxID(
	ctx context.Context,
	hash entities.TxID,
) (*entities.Transaction[entities.TransactionUtxoInput], error) {
	transactionResult, err := jsonrpc.
		NewRequest[[]any, data.Transaction]("getrawtransaction", []any{hash, 2}).
		Prepare(c.cfg.NodeURL, jsonrpc.WithContext(ctx)).
		Execute(nil)
	if err != nil {
		return nil, eris.Wrap(err, "GetRawTransaction: execute")
	}

	transaction := c.buildTransaction(transactionResult)

	return &transaction, nil
}

func (c *Client) buildTransaction(tx *data.Transaction) entities.Transaction[entities.TransactionUtxoInput] {
	inputs := make([]entities.TransactionUtxoInput, 0, len(tx.Vin))
	outputs := make([]entities.TransactionOutput, 0, len(tx.Vout))

	for _, vin := range tx.Vin {
		inputs = append(inputs, entities.TransactionUtxoInput{
			TxID: vin.TxID,
			N:    vin.Vout,
		})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, entities.TransactionOutput{
			N:       0,
			Value:   vout.Value,
			Ticker:  c.nativeTicker,
			Address: vout.ScriptPubKey.Address,
		})
	}

	transaction := entities.Transaction[entities.TransactionUtxoInput]{
		TxID:      tx.TxID,
		Inputs:    inputs,
		Outputs:   outputs,
		Fee:       tx.Fee,
		BlockHash: tx.BlockHash,
	}

	return transaction
}
