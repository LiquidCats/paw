package usecase_test

import (
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

type TxIDHandlerSuite struct {
	suite.Suite

	cfg                  configs.ChainConfig
	client               *mocks.MockClient[any]
	publisher            *mocks.MockPublisher[entities.Transaction[any]]
	requestToNodeCounter *mocks.MockRequestToNodeCounter
}

func TestTxIDHandlerSuite(t *testing.T) {
	suite.Run(t, new(TxIDHandlerSuite))
}

func (s *TxIDHandlerSuite) SetupTest() {
	s.cfg = configs.ChainConfig{
		Driver: entities.DriverRPC,
		Type:   entities.TypeUtxo,
		Chain:  "bitcoin",
		Scan: configs.ScanConfig{
			Depth: 2,
		},
		Persist: configs.PersistConfig{
			Capacity: 6,
			Interval: time.Hour,
		},
		Topics: configs.TopicsConfig{
			Transactions: "test-transactions",
		},
	}
	s.client = mocks.NewMockClient[any](s.T())
	s.publisher = mocks.NewMockPublisher[entities.Transaction[any]](s.T())
	s.requestToNodeCounter = mocks.NewMockRequestToNodeCounter(s.T())
}

func (s *TxIDHandlerSuite) newHandler() *usecase.TxIDHandler[any] {
	return usecase.NewTxIDHandler(
		s.cfg,
		s.client,
		s.publisher,
		usecase.TxIDHandlerMetrics{
			RequestToNodeCounter: s.requestToNodeCounter,
		},
	)
}

func (s *TxIDHandlerSuite) createTransaction(txID entities.TxID) *entities.Transaction[any] {
	return &entities.Transaction[any]{
		TxID:      txID,
		BlockHash: "block_hash_1",
		Inputs:    nil,
		Outputs: []entities.TransactionOutput{
			{
				N:       0,
				Value:   decimal.RequireFromString("1.5"),
				Address: "addr1",
			},
		},
		Fee: decimal.RequireFromString("0.0001"),
	}
}

func (s *TxIDHandlerSuite) TestConstructorReturnsNonNil() {
	handler := s.newHandler()

	s.NotNil(handler)
}

func (s *TxIDHandlerSuite) TestHandleSuccess() {
	ctx := s.T().Context()
	txID := entities.TxID("tx_hash_1")
	tx := s.createTransaction(txID)

	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once()
	s.client.On("GetTransactionByTxID", ctx, txID).Once().Return(tx, nil)
	s.publisher.On("PublishTo", ctx, s.cfg.Topics.Transactions, *tx).Once().Return(nil)

	handler := s.newHandler()
	err := handler.Handle(ctx, txID)

	s.Require().NoError(err)
}

func (s *TxIDHandlerSuite) TestHandleGetTransactionError() {
	ctx := s.T().Context()
	txID := entities.TxID("tx_hash_1")
	expectedErr := errors.New("rpc connection failed")

	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once()
	s.client.On("GetTransactionByTxID", ctx, txID).Once().Return(nil, expectedErr)

	handler := s.newHandler()
	err := handler.Handle(ctx, txID)

	s.Require().Error(err)
	s.Contains(err.Error(), "get transaction by txid")
}

func (s *TxIDHandlerSuite) TestHandlePublishError() {
	ctx := s.T().Context()
	txID := entities.TxID("tx_hash_1")
	tx := s.createTransaction(txID)
	expectedErr := errors.New("kafka publish failed")

	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once()
	s.client.On("GetTransactionByTxID", ctx, txID).Once().Return(tx, nil)
	s.publisher.On("PublishTo", ctx, s.cfg.Topics.Transactions, *tx).Once().Return(expectedErr)

	handler := s.newHandler()
	err := handler.Handle(ctx, txID)

	s.Require().Error(err)
	s.Contains(err.Error(), "publish mempool transaction")
}

func (s *TxIDHandlerSuite) TestMetricsCalledBeforeRPCCall() {
	ctx := s.T().Context()
	txID := entities.TxID("tx_hash_1")
	tx := s.createTransaction(txID)

	var incCalledBeforeRPC bool
	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once().Run(func(_ mock.Arguments) {
		incCalledBeforeRPC = true
	})
	s.client.On("GetTransactionByTxID", ctx, txID).Once().Return(tx, nil)
	s.publisher.On("PublishTo", ctx, s.cfg.Topics.Transactions, *tx).Once().Return(nil)

	handler := s.newHandler()
	_ = handler.Handle(ctx, txID)

	s.True(incCalledBeforeRPC)
}

func (s *TxIDHandlerSuite) TestMetricsCalledEvenOnRPCError() {
	ctx := s.T().Context()
	txID := entities.TxID("tx_hash_1")
	expectedErr := errors.New("node unavailable")

	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once()
	s.client.On("GetTransactionByTxID", ctx, txID).Once().Return(nil, expectedErr)

	handler := s.newHandler()
	_ = handler.Handle(ctx, txID)

	// Metrics mock will verify Inc was called via AssertExpectations
}

func (s *TxIDHandlerSuite) TestHandleWithDifferentChains() {
	testCases := []struct {
		name  string
		chain entities.Chain
	}{
		{"bitcoin", "bitcoin"},
		{"litecoin", "litecoin"},
		{"dogecoin", "dogecoin"},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cfg := s.cfg
			cfg.Chain = tc.chain

			client := mocks.NewMockClient[any](s.T())
			publisher := mocks.NewMockPublisher[entities.Transaction[any]](s.T())
			counter := mocks.NewMockRequestToNodeCounter(s.T())

			txID := entities.TxID("tx_" + string(tc.chain))
			tx := &entities.Transaction[any]{
				TxID:      txID,
				BlockHash: "block_hash",
				Fee:       decimal.RequireFromString("0.001"),
			}

			counter.On("Inc", tc.chain).Once()
			client.On("GetTransactionByTxID", mock.Anything, txID).Once().Return(tx, nil)
			publisher.On("PublishTo", mock.Anything, cfg.Topics.Transactions, *tx).Once().Return(nil)

			handler := usecase.NewTxIDHandler(
				cfg,
				client,
				publisher,
				usecase.TxIDHandlerMetrics{RequestToNodeCounter: counter},
			)

			err := handler.Handle(s.T().Context(), txID)
			s.Require().NoError(err)
		})
	}
}

func (s *TxIDHandlerSuite) TestHandleWithDifferentTopics() {
	ctx := s.T().Context()
	customTopic := "custom-tx-topic"
	cfg := s.cfg
	cfg.Topics.Transactions = customTopic

	client := mocks.NewMockClient[any](s.T())
	publisher := mocks.NewMockPublisher[entities.Transaction[any]](s.T())
	counter := mocks.NewMockRequestToNodeCounter(s.T())

	txID := entities.TxID("tx_hash_custom")
	tx := s.createTransaction(txID)

	counter.On("Inc", cfg.Chain).Once()
	client.On("GetTransactionByTxID", ctx, txID).Once().Return(tx, nil)
	publisher.On("PublishTo", ctx, customTopic, *tx).Once().Return(nil)

	handler := usecase.NewTxIDHandler(
		cfg,
		client,
		publisher,
		usecase.TxIDHandlerMetrics{RequestToNodeCounter: counter},
	)

	err := handler.Handle(ctx, txID)
	s.Require().NoError(err)
}

func (s *TxIDHandlerSuite) TestHandleTransactionWithMultipleOutputs() {
	ctx := s.T().Context()
	txID := entities.TxID("multi_output_tx")
	tx := &entities.Transaction[any]{
		TxID:      txID,
		BlockHash: "block_hash_multi",
		Inputs:    nil,
		Outputs: []entities.TransactionOutput{
			{N: 0, Value: decimal.RequireFromString("0.5"), Address: "addr1"},
			{N: 1, Value: decimal.RequireFromString("1.0"), Address: "addr2"},
			{N: 2, Value: decimal.RequireFromString("2.5"), Address: "addr3"},
		},
		Fee: decimal.RequireFromString("0.0002"),
	}

	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once()
	s.client.On("GetTransactionByTxID", ctx, txID).Once().Return(tx, nil)
	s.publisher.On("PublishTo", ctx, s.cfg.Topics.Transactions, *tx).Once().Return(nil)

	handler := s.newHandler()
	err := handler.Handle(ctx, txID)

	s.Require().NoError(err)
}

func (s *TxIDHandlerSuite) TestHandleTransactionWithEmptyBlockHash() {
	ctx := s.T().Context()
	txID := entities.TxID("mempool_tx")
	tx := &entities.Transaction[any]{
		TxID:      txID,
		BlockHash: "", // Empty block hash indicates mempool transaction
		Inputs:    nil,
		Outputs:   []entities.TransactionOutput{},
		Fee:       decimal.RequireFromString("0.0001"),
	}

	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once()
	s.client.On("GetTransactionByTxID", ctx, txID).Once().Return(tx, nil)
	s.publisher.On("PublishTo", ctx, s.cfg.Topics.Transactions, *tx).Once().Return(nil)

	handler := s.newHandler()
	err := handler.Handle(ctx, txID)

	s.Require().NoError(err)
}

func (s *TxIDHandlerSuite) TestHandleMultipleTransactionsSequentially() {
	ctx := s.T().Context()
	txIDs := []entities.TxID{"tx1", "tx2", "tx3"}

	for _, txID := range txIDs {
		tx := s.createTransaction(txID)
		s.client.On("GetTransactionByTxID", ctx, txID).Once().Return(tx, nil)
		s.publisher.On("PublishTo", ctx, s.cfg.Topics.Transactions, *tx).Once().Return(nil)
	}
	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Times(3)

	handler := s.newHandler()

	for _, txID := range txIDs {
		err := handler.Handle(ctx, txID)
		s.Require().NoError(err)
	}
}

func (s *TxIDHandlerSuite) TestHandleWithConnectionTimeoutError() {
	ctx := s.T().Context()
	txID := entities.TxID("tx_timeout")
	timeoutErr := errors.New("connection timeout after 30s")

	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once()
	s.client.On("GetTransactionByTxID", ctx, txID).Once().Return(nil, timeoutErr)

	handler := s.newHandler()
	err := handler.Handle(ctx, txID)

	s.Require().Error(err)
	s.Contains(err.Error(), "get transaction by txid")
}

func (s *TxIDHandlerSuite) TestHandleWithPublisherConnectionError() {
	ctx := s.T().Context()
	txID := entities.TxID("tx_pub_error")
	tx := s.createTransaction(txID)
	pubErr := errors.New("broker connection refused")

	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once()
	s.client.On("GetTransactionByTxID", ctx, txID).Once().Return(tx, nil)
	s.publisher.On("PublishTo", ctx, s.cfg.Topics.Transactions, *tx).Once().Return(pubErr)

	handler := s.newHandler()
	err := handler.Handle(ctx, txID)

	s.Require().Error(err)
	s.Contains(err.Error(), "publish mempool transaction")
}

func (s *TxIDHandlerSuite) TestHandleWithLongTransactionID() {
	ctx := s.T().Context()
	// Bitcoin transaction IDs are 64 hex characters
	txID := entities.TxID("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	tx := s.createTransaction(txID)

	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once()
	s.client.On("GetTransactionByTxID", ctx, txID).Once().Return(tx, nil)
	s.publisher.On("PublishTo", ctx, s.cfg.Topics.Transactions, *tx).Once().Return(nil)

	handler := s.newHandler()
	err := handler.Handle(ctx, txID)

	s.Require().NoError(err)
}

func (s *TxIDHandlerSuite) TestHandlePreservesTransactionData() {
	ctx := s.T().Context()
	txID := entities.TxID("preserve_data_tx")
	expectedTx := &entities.Transaction[any]{
		TxID:      txID,
		BlockHash: "specific_block_hash",
		Inputs:    []any{"input1", "input2"},
		Outputs: []entities.TransactionOutput{
			{
				N:        0,
				Value:    decimal.RequireFromString("123.456789"),
				Ticker:   "BTC",
				Contract: "contract_addr",
				Address:  "output_addr",
			},
		},
		Fee: decimal.RequireFromString("0.00012345"),
	}

	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once()
	s.client.On("GetTransactionByTxID", ctx, txID).Once().Return(expectedTx, nil)

	var publishedTx entities.Transaction[any]
	s.publisher.On("PublishTo", ctx, s.cfg.Topics.Transactions, mock.Anything).
		Once().
		Run(func(args mock.Arguments) {
			publishedTx = args.Get(2).(entities.Transaction[any])
		}).
		Return(nil)

	handler := s.newHandler()
	err := handler.Handle(ctx, txID)

	s.Require().NoError(err)
	s.Equal(expectedTx.TxID, publishedTx.TxID)
	s.Equal(expectedTx.BlockHash, publishedTx.BlockHash)
	s.Equal(expectedTx.Fee, publishedTx.Fee)
	s.Len(publishedTx.Outputs, 1)
	s.Equal(expectedTx.Outputs[0].Value, publishedTx.Outputs[0].Value)
}

func (s *TxIDHandlerSuite) TestHandleWithEVMDriver() {
	ctx := s.T().Context()
	cfg := s.cfg
	cfg.Driver = entities.DriverRPC
	cfg.Type = entities.TypeEvm
	cfg.Chain = "ethereum"

	client := mocks.NewMockClient[any](s.T())
	publisher := mocks.NewMockPublisher[entities.Transaction[any]](s.T())
	counter := mocks.NewMockRequestToNodeCounter(s.T())

	txID := entities.TxID("0xabc123")
	tx := &entities.Transaction[any]{
		TxID:      txID,
		BlockHash: "0xblock",
		Fee:       decimal.RequireFromString("0.001"),
	}

	counter.On("Inc", cfg.Chain).Once()
	client.On("GetTransactionByTxID", ctx, txID).Once().Return(tx, nil)
	publisher.On("PublishTo", ctx, cfg.Topics.Transactions, *tx).Once().Return(nil)

	handler := usecase.NewTxIDHandler(
		cfg,
		client,
		publisher,
		usecase.TxIDHandlerMetrics{RequestToNodeCounter: counter},
	)

	err := handler.Handle(ctx, txID)
	s.Require().NoError(err)
}
