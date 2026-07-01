package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/LiquidCats/paw/services/watcher/configs"
	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
	"github.com/LiquidCats/paw/services/watcher/internal/app/usecase"
	"github.com/LiquidCats/paw/services/watcher/test/mocks"
	"github.com/stretchr/testify/suite"
)

type BlocksPersisterSuite struct {
	suite.Suite

	ctx context.Context
	cfg configs.ChainConfig

	mockDB    *mocks.MockStateDB
	mockState *mocks.MockSliceState[entities.BlockHash]
}

func TestBlocksPersisterSuite(t *testing.T) {
	suite.Run(t, new(BlocksPersisterSuite))
}

func (s *BlocksPersisterSuite) SetupTest() {
	s.ctx = context.Background()
	s.cfg = configs.ChainConfig{
		Driver:  "rpc",
		Type:    "evm",
		Chain:   "ethereum",
		Persist: configs.PersistConfig{},
		Scan:    configs.ScanConfig{},
		Workers: configs.WorkersConfig{},
		RPC:     configs.RPCConfig{},
		Topics:  configs.TopicsConfig{},
	}

	s.mockDB = mocks.NewMockStateDB(s.T())
	s.mockState = mocks.NewMockSliceState[entities.BlockHash](s.T())
}

func (s *BlocksPersisterSuite) newPersister() *usecase.BlocksPersister {
	return usecase.NewBlocksPersister(s.cfg, s.mockDB, s.mockState)
}

func (s *BlocksPersisterSuite) TestSuccessfulPersistence() {
	blockHashes := []entities.BlockHash{"hash1", "hash2", "hash3"}

	s.mockState.EXPECT().Get().Return(blockHashes)
	s.mockDB.EXPECT().SetBlockState(s.ctx, s.cfg.Chain, blockHashes).Return(nil)

	persister := s.newPersister()
	err := persister.Handle(s.ctx)
	s.Require().NoError(err)

	s.mockState.AssertExpectations(s.T())
	s.mockDB.AssertExpectations(s.T())
}

func (s *BlocksPersisterSuite) TestSuccessfulPersistenceWithSingleBlock() {
	blockHashes := []entities.BlockHash{"single_hash"}

	s.mockState.EXPECT().Get().Return(blockHashes)
	s.mockDB.EXPECT().SetBlockState(s.ctx, s.cfg.Chain, blockHashes).Return(nil)

	persister := s.newPersister()
	err := persister.Handle(s.ctx)
	s.Require().NoError(err)

	s.mockState.AssertExpectations(s.T())
	s.mockDB.AssertExpectations(s.T())
}

func (s *BlocksPersisterSuite) TestEmptyState() {
	blockHashes := []entities.BlockHash{}

	s.mockState.EXPECT().Get().Return(blockHashes)
	s.mockDB.EXPECT().SetBlockState(s.ctx, s.cfg.Chain, blockHashes).Return(nil)

	persister := s.newPersister()
	err := persister.Handle(s.ctx)
	s.Require().NoError(err)

	s.mockState.AssertExpectations(s.T())
	s.mockDB.AssertExpectations(s.T())
}

func (s *BlocksPersisterSuite) TestNilState() {
	var blockHashes []entities.BlockHash

	s.mockState.EXPECT().Get().Return(blockHashes)
	s.mockDB.EXPECT().SetBlockState(s.ctx, s.cfg.Chain, blockHashes).Return(nil)

	persister := s.newPersister()
	err := persister.Handle(s.ctx)
	s.Require().NoError(err)

	s.mockState.AssertExpectations(s.T())
	s.mockDB.AssertExpectations(s.T())
}

func (s *BlocksPersisterSuite) TestDatabaseError() {
	blockHashes := []entities.BlockHash{"hash1", "hash2"}
	dbErr := errors.New("database connection failed")

	s.mockState.EXPECT().Get().Return(blockHashes)
	s.mockDB.EXPECT().SetBlockState(s.ctx, s.cfg.Chain, blockHashes).Return(dbErr)

	persister := s.newPersister()
	err := persister.Handle(s.ctx)
	s.Require().Error(err)
	s.Contains(err.Error(), "set blocks state")

	s.mockState.AssertExpectations(s.T())
	s.mockDB.AssertExpectations(s.T())
}

func (s *BlocksPersisterSuite) TestDatabaseTimeoutError() {
	blockHashes := []entities.BlockHash{"hash1"}
	dbErr := context.DeadlineExceeded

	s.mockState.EXPECT().Get().Return(blockHashes)
	s.mockDB.EXPECT().SetBlockState(s.ctx, s.cfg.Chain, blockHashes).Return(dbErr)

	persister := s.newPersister()
	err := persister.Handle(s.ctx)
	s.Require().Error(err)
	s.Contains(err.Error(), "set blocks state")

	s.mockState.AssertExpectations(s.T())
	s.mockDB.AssertExpectations(s.T())
}

func (s *BlocksPersisterSuite) TestDifferentChainConfig() {
	s.cfg.Chain = "bitcoin"
	s.cfg.Type = "utxo"
	s.cfg.Driver = "bitcoin-rpc"

	blockHashes := []entities.BlockHash{"btc_hash_1", "btc_hash_2"}

	s.mockState.EXPECT().Get().Return(blockHashes)
	s.mockDB.EXPECT().SetBlockState(s.ctx, entities.Chain("bitcoin"), blockHashes).Return(nil)

	persister := s.newPersister()
	err := persister.Handle(s.ctx)
	s.Require().NoError(err)

	s.mockState.AssertExpectations(s.T())
	s.mockDB.AssertExpectations(s.T())
}
