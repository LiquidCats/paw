package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/LiquidCats/paw/watcher/configs"
	"github.com/LiquidCats/paw/watcher/internal/app/domain/entities"
	"github.com/LiquidCats/paw/watcher/internal/app/usecase"
	"github.com/LiquidCats/paw/watcher/test/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MempoolDiggerSuite struct {
	suite.Suite

	cfg                  configs.ChainConfig
	client               *mocks.MockClient[any]
	requestToNodeCounter *mocks.MockRequestToNodeCounter
}

func TestMempoolDiggerSuite(t *testing.T) {
	suite.Run(t, new(MempoolDiggerSuite))
}

func (s *MempoolDiggerSuite) SetupTest() {
	s.cfg = configs.ChainConfig{
		Driver: entities.DriverRPC,
		Type:   entities.TypeUtxo,
		Chain:  "bitcoin",
		Persist: configs.PersistConfig{
			Interval: time.Hour,
		},
	}
	s.client = mocks.NewMockClient[any](s.T())
	s.requestToNodeCounter = mocks.NewMockRequestToNodeCounter(s.T())
}

func (s *MempoolDiggerSuite) newDigger(txCh chan entities.TxID, oldMempool []entities.TxID) *usecase.MempoolDigger[any] {
	return usecase.NewMempoolDigger(
		s.cfg,
		s.client,
		txCh,
		oldMempool,
		usecase.MempoolJobMetrics{
			RequestToNodeCounter: s.requestToNodeCounter,
		},
	)
}

func (s *MempoolDiggerSuite) collectTxIDs(txCh chan entities.TxID) []entities.TxID {
	var result []entities.TxID
	for {
		select {
		case txID := <-txCh:
			result = append(result, txID)
		default:
			return result
		}
	}
}

func (s *MempoolDiggerSuite) TestNewTransactionsWithEmptyOldMempool() {
	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.client.On("GetMempool", mock.Anything).Return([]entities.TxID{"tx1", "tx2", "tx3"}, nil)

	txCh := make(chan entities.TxID, 100)
	uc := s.newDigger(txCh, []entities.TxID{})

	err := uc.Handle(s.T().Context())

	s.Require().NoError(err)
	s.Equal([]entities.TxID{"tx1", "tx2", "tx3"}, s.collectTxIDs(txCh))
}

func (s *MempoolDiggerSuite) TestEmptyNewMempoolReturnsEarly() {
	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.client.On("GetMempool", mock.Anything).Return([]entities.TxID{}, nil)

	txCh := make(chan entities.TxID, 100)
	uc := s.newDigger(txCh, []entities.TxID{"tx1"})

	err := uc.Handle(s.T().Context())

	s.Require().NoError(err)
	s.Nil(s.collectTxIDs(txCh))
}

func (s *MempoolDiggerSuite) TestNilNewMempoolReturnsEarly() {
	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.client.On("GetMempool", mock.Anything).Return(nil, nil)

	txCh := make(chan entities.TxID, 100)
	uc := s.newDigger(txCh, []entities.TxID{"tx1"})

	err := uc.Handle(s.T().Context())

	s.Require().NoError(err)
	s.Nil(s.collectTxIDs(txCh))
}

func (s *MempoolDiggerSuite) TestGetMempoolError() {
	expectedErr := errors.New("rpc connection failed")
	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.client.On("GetMempool", mock.Anything).Return(nil, expectedErr)

	txCh := make(chan entities.TxID, 100)
	uc := s.newDigger(txCh, []entities.TxID{})

	err := uc.Handle(s.T().Context())

	s.Require().Error(err)
	s.Contains(err.Error(), "get new mempool")
	s.Nil(s.collectTxIDs(txCh))
}

func (s *MempoolDiggerSuite) TestOnlyNewTransactionsSentWhenOldMempoolHasEntries() {
	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.client.On("GetMempool", mock.Anything).Return([]entities.TxID{"tx1", "tx2", "tx3", "tx4"}, nil)

	txCh := make(chan entities.TxID, 100)
	uc := s.newDigger(txCh, []entities.TxID{"tx1", "tx2"})

	err := uc.Handle(s.T().Context())

	s.Require().NoError(err)
	s.Equal([]entities.TxID{"tx3", "tx4"}, s.collectTxIDs(txCh))
}

func (s *MempoolDiggerSuite) TestNoNewTransactionsWhenAllExistInOldMempool() {
	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.client.On("GetMempool", mock.Anything).Return([]entities.TxID{"tx1", "tx2"}, nil)

	txCh := make(chan entities.TxID, 100)
	uc := s.newDigger(txCh, []entities.TxID{"tx1", "tx2", "tx3"})

	err := uc.Handle(s.T().Context())

	s.Require().NoError(err)
	s.Nil(s.collectTxIDs(txCh))
}

func (s *MempoolDiggerSuite) TestPartialOverlapWithNewAndRemovedTransactions() {
	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.client.On("GetMempool", mock.Anything).Return([]entities.TxID{"tx2", "tx4", "tx5"}, nil)

	txCh := make(chan entities.TxID, 100)
	uc := s.newDigger(txCh, []entities.TxID{"tx1", "tx2", "tx3"})

	err := uc.Handle(s.T().Context())

	s.Require().NoError(err)
	s.Equal([]entities.TxID{"tx4", "tx5"}, s.collectTxIDs(txCh))
}

func (s *MempoolDiggerSuite) TestSingleNewTransaction() {
	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.client.On("GetMempool", mock.Anything).Return([]entities.TxID{"tx1"}, nil)

	txCh := make(chan entities.TxID, 100)
	uc := s.newDigger(txCh, []entities.TxID{})

	err := uc.Handle(s.T().Context())

	s.Require().NoError(err)
	s.Equal([]entities.TxID{"tx1"}, s.collectTxIDs(txCh))
}

func (s *MempoolDiggerSuite) TestCompletelyDifferentMempools() {
	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.client.On("GetMempool", mock.Anything).Return([]entities.TxID{"tx3", "tx4"}, nil)

	txCh := make(chan entities.TxID, 100)
	uc := s.newDigger(txCh, []entities.TxID{"tx1", "tx2"})

	err := uc.Handle(s.T().Context())

	s.Require().NoError(err)
	s.Equal([]entities.TxID{"tx3", "tx4"}, s.collectTxIDs(txCh))
}

func (s *MempoolDiggerSuite) TestUpdatesOldMempoolAcrossMultipleCalls() {
	// First call - returns tx1, tx2
	// Second call - returns tx2, tx3 (tx1 removed, tx3 added)
	// Third call - returns tx3, tx4, tx5 (tx2 removed, tx4, tx5 added)
	s.client.
		On("GetMempool", mock.Anything).
		Return([]entities.TxID{"tx1", "tx2"}, nil).Once()
	s.client.
		On("GetMempool", mock.Anything).
		Return([]entities.TxID{"tx2", "tx3"}, nil).Once()
	s.client.
		On("GetMempool", mock.Anything).
		Return([]entities.TxID{"tx3", "tx4", "tx5"}, nil).Once()

	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Times(3).Return(nil)

	txCh := make(chan entities.TxID, 100)
	uc := s.newDigger(txCh, []entities.TxID{})

	// First Handle call
	err := uc.Handle(s.T().Context())
	s.Require().NoError(err)

	// Second Handle call - should only send tx3 (tx2 was already in old mempool)
	err = uc.Handle(s.T().Context())
	s.Require().NoError(err)

	// Third Handle call - should only send tx4, tx5 (tx3 was already in old mempool)
	err = uc.Handle(s.T().Context())
	s.Require().NoError(err)

	// First call: tx1, tx2 (all new)
	// Second call: tx3 (only new one)
	// Third call: tx4, tx5 (only new ones)
	expected := []entities.TxID{"tx1", "tx2", "tx3", "tx4", "tx5"}
	s.Equal(expected, s.collectTxIDs(txCh))
}

func (s *MempoolDiggerSuite) TestMetricsCalledOnError() {
	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.client.On("GetMempool", mock.Anything).Return(nil, errors.New("connection timeout"))

	txCh := make(chan entities.TxID, 100)
	uc := s.newDigger(txCh, []entities.TxID{})

	err := uc.Handle(s.T().Context())

	s.Require().Error(err)
	// The mock will verify that Inc was called via AssertExpectations
}

func (s *MempoolDiggerSuite) TestLargeMempool() {
	const mempoolSize = 1000
	newMempool := make([]entities.TxID, mempoolSize)
	for i := range mempoolSize {
		newMempool[i] = entities.TxID("tx" + string(rune('a'+i%26)) + string(rune('0'+i/26%10)))
	}

	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.client.On("GetMempool", mock.Anything).Return(newMempool, nil)

	txCh := make(chan entities.TxID, mempoolSize)
	uc := s.newDigger(txCh, []entities.TxID{})

	err := uc.Handle(s.T().Context())

	s.Require().NoError(err)
	s.Len(s.collectTxIDs(txCh), mempoolSize)
}

func (s *MempoolDiggerSuite) TestPreservesTransactionOrder() {
	newMempool := []entities.TxID{"tx5", "tx3", "tx1", "tx4", "tx2"}

	s.requestToNodeCounter.On("Inc", s.cfg.Chain).Once().Return(nil)
	s.client.On("GetMempool", mock.Anything).Return(newMempool, nil)

	txCh := make(chan entities.TxID, 100)
	uc := s.newDigger(txCh, []entities.TxID{})

	err := uc.Handle(s.T().Context())

	s.Require().NoError(err)
	s.Equal(newMempool, s.collectTxIDs(txCh))
}

func (s *MempoolDiggerSuite) TestConstructorReturnsNonNil() {
	txCh := make(chan entities.TxID, 1)
	uc := s.newDigger(txCh, []entities.TxID{"tx1", "tx2"})

	s.NotNil(uc)
}
