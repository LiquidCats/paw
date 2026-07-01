package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LiquidCats/paw/services/watcher/configs"
	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
	"github.com/LiquidCats/paw/services/watcher/internal/app/usecase"
	"github.com/LiquidCats/paw/services/watcher/test/mocks"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type BlockTransactionsHandlerSuite struct {
	suite.Suite

	ctx context.Context
	cfg configs.ChainConfig

	blockHash0 entities.BlockHash
	blockHash1 entities.BlockHash

	tx1 *entities.Transaction[entities.TransactionUtxoInput]
	tx2 *entities.Transaction[entities.TransactionUtxoInput]

	blockHeader              entities.BlockHeader
	block                    *entities.Block
	blockWithOneTransaction  *entities.BlockWithTransactions[entities.TransactionUtxoInput]
	blockWithTwoTransactions *entities.BlockWithTransactions[entities.TransactionUtxoInput]
	blockWithNoTransactions  *entities.BlockWithTransactions[entities.TransactionUtxoInput]

	mockTxPub                *mocks.MockPublisher[entities.Transaction[entities.TransactionUtxoInput]]
	mockBlockPub             *mocks.MockPublisher[entities.Block]
	mockRPC                  *mocks.MockClient[entities.TransactionUtxoInput]
	mockFlight               *mocks.MockMapState[entities.BlockHash, bool]
	mockRequestToNodeCounter *mocks.MockRequestToNodeCounter
}

func TestBlockTransactionsHandlerSuite(t *testing.T) {
	suite.Run(t, new(BlockTransactionsHandlerSuite))
}

func (s *BlockTransactionsHandlerSuite) SetupTest() {
	s.ctx = context.Background()
	s.cfg = configs.ChainConfig{
		Driver: "rpc",
		Type:   "utxo",
		ISO:    "BTC",
		Chain:  "bitcoin",
		Persist: configs.PersistConfig{
			Capacity: 3,
			Interval: 1 * time.Second,
		},
		Scan: configs.ScanConfig{
			Depth:    1,
			Interval: 1 * time.Second,
		},
		Workers: configs.WorkersConfig{
			TxIDWorkerCount:              3,
			BlockTransactionsWorkerCount: 3,
		},
		RPC: configs.RPCConfig{
			NodeURL: "rpc.test",
		},
		Topics: configs.TopicsConfig{
			Transactions: "test-transactions",
			Blocks:       "test-blocks",
		},
	}

	s.blockHash0 = entities.BlockHash("blockhahs0")
	s.blockHash1 = entities.BlockHash("blockhahs1")

	s.tx1 = &entities.Transaction[entities.TransactionUtxoInput]{
		TxID: entities.TxID("tx1"),
		Inputs: []entities.TransactionUtxoInput{
			{
				TxID: "txout1",
				N:    0,
			},
		},
		Outputs: []entities.TransactionOutput{
			{
				N:       0,
				Value:   decimal.NewFromFloat(0.7),
				Ticker:  "BTC",
				Address: "test_addres_1",
			},
		},
		Fee:       decimal.NewFromFloat(0.11),
		BlockHash: s.blockHash1,
	}

	s.tx2 = &entities.Transaction[entities.TransactionUtxoInput]{
		TxID: entities.TxID("tx2"),
		Inputs: []entities.TransactionUtxoInput{
			{
				TxID: "txout1",
				N:    0,
			},
		},
		Outputs: []entities.TransactionOutput{
			{
				N:       0,
				Value:   decimal.NewFromFloat(1.01),
				Ticker:  "BTC",
				Address: "test_addres_1",
			},
		},
		Fee:       decimal.NewFromFloat(0.10),
		BlockHash: s.blockHash1,
	}

	s.blockHeader = entities.BlockHeader{
		Height:   1,
		Hash:     s.blockHash1,
		PrevHash: s.blockHash0,
	}

	s.block = &entities.Block{
		BlockHeader: s.blockHeader,
		Transactions: []entities.TxID{
			s.tx1.TxID,
			s.tx2.TxID,
		},
	}

	s.blockWithOneTransaction = &entities.BlockWithTransactions[entities.TransactionUtxoInput]{
		BlockHeader: s.blockHeader,
		Transactions: []entities.Transaction[entities.TransactionUtxoInput]{
			*s.tx1,
		},
	}

	s.blockWithNoTransactions = &entities.BlockWithTransactions[entities.TransactionUtxoInput]{
		BlockHeader:  s.blockHeader,
		Transactions: []entities.Transaction[entities.TransactionUtxoInput]{},
	}

	s.blockWithTwoTransactions = &entities.BlockWithTransactions[entities.TransactionUtxoInput]{
		BlockHeader: s.blockHeader,
		Transactions: []entities.Transaction[entities.TransactionUtxoInput]{
			*s.tx1,
			*s.tx2,
		},
	}

	s.mockTxPub = mocks.NewMockPublisher[entities.Transaction[entities.TransactionUtxoInput]](s.T())
	s.mockBlockPub = mocks.NewMockPublisher[entities.Block](s.T())
	s.mockRPC = mocks.NewMockClient[entities.TransactionUtxoInput](s.T())
	s.mockFlight = mocks.NewMockMapState[entities.BlockHash, bool](s.T())
	s.mockRequestToNodeCounter = mocks.NewMockRequestToNodeCounter(s.T())
}

func (s *BlockTransactionsHandlerSuite) newHandler() *usecase.BlockTransactionsHandler[entities.TransactionUtxoInput] {
	return usecase.NewBlockTransactionsHandler(
		s.cfg,
		s.mockRPC,
		s.mockTxPub,
		s.mockBlockPub,
		s.mockFlight,
		usecase.BlockTransactionsHandlerMetrics{
			RequestToNodeCounter: s.mockRequestToNodeCounter,
		},
	)
}

func (s *BlockTransactionsHandlerSuite) TestSuccessfulHandlingWithOneTransaction() {
	s.mockRPC.On("GetBlockByHashWithTransactions", mock.Anything, s.blockHeader.Hash).Once().Return(s.blockWithOneTransaction, nil)
	s.mockTxPub.On("PublishTo", mock.Anything, s.cfg.Topics.Transactions, *s.tx1).Once().Return(nil)
	s.mockBlockPub.On("PublishTo", mock.Anything, s.cfg.Topics.Blocks, *s.block).Once().Return(nil)
	s.mockRequestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.mockFlight.On("Del", s.block.Hash).Once().Return()

	handler := s.newHandler()
	err := handler.Handle(s.ctx, s.block)
	s.Require().NoError(err)

	s.mockFlight.AssertExpectations(s.T())
}

func (s *BlockTransactionsHandlerSuite) TestSuccessfulHandlingWithMultipleTransactions() {
	s.mockRPC.On("GetBlockByHashWithTransactions", mock.Anything, s.blockHeader.Hash).Once().Return(s.blockWithTwoTransactions, nil)
	s.mockTxPub.On("PublishTo", mock.Anything, s.cfg.Topics.Transactions, *s.tx1).Once().Return(nil)
	s.mockTxPub.On("PublishTo", mock.Anything, s.cfg.Topics.Transactions, *s.tx2).Once().Return(nil)
	s.mockBlockPub.On("PublishTo", mock.Anything, s.cfg.Topics.Blocks, *s.block).Once().Return(nil)
	s.mockRequestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.mockFlight.On("Del", s.block.Hash).Once()

	handler := s.newHandler()
	err := handler.Handle(s.ctx, s.block)
	s.Require().NoError(err)

	s.mockFlight.AssertExpectations(s.T())
}

func (s *BlockTransactionsHandlerSuite) TestErrorWhenGettingBlockByHash() {
	s.mockRPC.On("GetBlockByHashWithTransactions", mock.Anything, s.block.Hash).Once().Return(nil, errors.New("rpc error"))
	s.mockFlight.On("Del", s.block.Hash).Once()

	handler := s.newHandler()
	err := handler.Handle(s.ctx, s.block)
	s.Require().Error(err)

	s.mockRPC.AssertExpectations(s.T())
	s.mockRequestToNodeCounter.AssertExpectations(s.T())
	s.mockFlight.AssertExpectations(s.T())
}

func (s *BlockTransactionsHandlerSuite) TestErrorWhenPublishingTransaction() {
	s.mockRPC.On("GetBlockByHashWithTransactions", mock.Anything, s.block.Hash).Once().Return(s.blockWithOneTransaction, nil)
	s.mockTxPub.On("PublishTo", mock.Anything, s.cfg.Topics.Transactions, *s.tx1).Once().Return(errors.New("publish error"))
	s.mockRequestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.mockFlight.On("Del", s.block.Hash).Once()

	handler := s.newHandler()
	err := handler.Handle(s.ctx, s.block)
	s.Require().Error(err)
	s.Contains(err.Error(), "publish transaction")
}

func (s *BlockTransactionsHandlerSuite) TestErrorWhenPublishingBlock() {
	s.mockRPC.On("GetBlockByHashWithTransactions", mock.Anything, s.block.Hash).Once().Return(s.blockWithOneTransaction, nil)
	s.mockTxPub.On("PublishTo", mock.Anything, s.cfg.Topics.Transactions, *s.tx1).Once().Return(nil)
	s.mockBlockPub.On("PublishTo", mock.Anything, s.cfg.Topics.Blocks, *s.block).Once().Return(errors.New("publish error"))
	s.mockRequestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.mockFlight.On("Del", s.block.Hash).Once()

	handler := s.newHandler()
	err := handler.Handle(s.ctx, s.block)
	s.Require().Error(err)

	s.mockRPC.AssertExpectations(s.T())
	s.mockTxPub.AssertExpectations(s.T())
	s.mockBlockPub.AssertExpectations(s.T())
	s.mockRequestToNodeCounter.AssertExpectations(s.T())
	s.mockFlight.AssertExpectations(s.T())
}

func (s *BlockTransactionsHandlerSuite) TestEmptyTransactions() {
	s.mockRPC.On("GetBlockByHashWithTransactions", mock.Anything, s.block.Hash).Once().Return(s.blockWithNoTransactions, nil)
	s.mockBlockPub.On("PublishTo", mock.Anything, s.cfg.Topics.Blocks, *s.block).Once().Return(nil)
	s.mockRequestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.mockFlight.On("Del", s.block.Hash).Once()

	handler := s.newHandler()
	err := handler.Handle(s.ctx, s.block)
	s.Require().NoError(err)

	s.mockRPC.AssertExpectations(s.T())
	s.mockBlockPub.AssertExpectations(s.T())
	s.mockRequestToNodeCounter.AssertExpectations(s.T())
	s.mockFlight.AssertExpectations(s.T())
}

func (s *BlockTransactionsHandlerSuite) TestContextCancellationDuringRpcCall() {
	ctx, cancel := context.WithCancel(s.ctx)
	cancel()

	s.mockRPC.On("GetBlockByHashWithTransactions", mock.Anything, s.block.Hash).Once().Return(nil, context.Canceled)
	s.mockFlight.On("Del", s.block.Hash).Once()

	handler := s.newHandler()
	err := handler.Handle(ctx, s.block)
	s.Require().Error(err)

	s.mockFlight.AssertExpectations(s.T())
}

func (s *BlockTransactionsHandlerSuite) TestNilBlockTransactionsEdgeCase() {
	blockWithNilTransactions := &entities.BlockWithTransactions[entities.TransactionUtxoInput]{
		BlockHeader:  s.blockHeader,
		Transactions: nil,
	}

	s.mockRPC.On("GetBlockByHashWithTransactions", mock.Anything, s.block.Hash).Once().Return(blockWithNilTransactions, nil)
	s.mockBlockPub.On("PublishTo", mock.Anything, s.cfg.Topics.Blocks, *s.block).Once().Return(nil)
	s.mockRequestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.mockFlight.On("Del", s.block.Hash).Once()

	handler := s.newHandler()
	err := handler.Handle(s.ctx, s.block)
	s.Require().NoError(err)

	s.mockFlight.AssertExpectations(s.T())
}

func (s *BlockTransactionsHandlerSuite) TestSuccessfulHandlingWithTransactionThatHasNoOutputs() {
	txWithoutOutputs := &entities.Transaction[entities.TransactionUtxoInput]{
		TxID: entities.TxID("tx_no_outputs"),
		Inputs: []entities.TransactionUtxoInput{
			{
				TxID: "txout1",
				N:    0,
			},
		},
		Outputs:   []entities.TransactionOutput{},
		Fee:       decimal.NewFromFloat(0.1),
		BlockHash: s.blockHash1,
	}

	blockWithTxNoOutputs := &entities.BlockWithTransactions[entities.TransactionUtxoInput]{
		BlockHeader: s.blockHeader,
		Transactions: []entities.Transaction[entities.TransactionUtxoInput]{
			*txWithoutOutputs,
		},
	}

	s.mockRPC.On("GetBlockByHashWithTransactions", mock.Anything, s.block.Hash).Once().Return(blockWithTxNoOutputs, nil)
	s.mockTxPub.On("PublishTo", mock.Anything, s.cfg.Topics.Transactions, *txWithoutOutputs).Once().Return(nil)
	s.mockBlockPub.On("PublishTo", mock.Anything, s.cfg.Topics.Blocks, *s.block).Once().Return(nil)
	s.mockRequestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.mockFlight.On("Del", s.block.Hash).Once()

	handler := s.newHandler()
	err := handler.Handle(s.ctx, s.block)
	s.Require().NoError(err)

	s.mockFlight.AssertExpectations(s.T())
}

func (s *BlockTransactionsHandlerSuite) TestMultipleTransactionsWithSomePublishErrors() {
	s.mockRPC.On("GetBlockByHashWithTransactions", mock.Anything, s.block.Hash).Once().Return(s.blockWithTwoTransactions, nil)
	s.mockTxPub.On("PublishTo", mock.Anything, s.cfg.Topics.Transactions, *s.tx1).Once().Return(nil)
	s.mockTxPub.On("PublishTo", mock.Anything, s.cfg.Topics.Transactions, *s.tx2).Once().Return(errors.New("publish error"))
	s.mockRequestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.mockFlight.On("Del", s.block.Hash).Once()

	handler := s.newHandler()
	err := handler.Handle(s.ctx, s.block)
	s.Require().Error(err)
	s.Contains(err.Error(), "publish error")
}
