package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/LiquidCats/paw/watcher/configs"
	"github.com/LiquidCats/paw/watcher/internal/app/domain/entities"
	"github.com/LiquidCats/paw/watcher/internal/app/usecase"
	"github.com/LiquidCats/paw/watcher/test/mocks"
	"github.com/stretchr/testify/suite"
)

type BlockchainDiggerSuite struct {
	suite.Suite

	ctx context.Context
	cfg configs.ChainConfig

	blockHash1 entities.BlockHash
	blockHash2 entities.BlockHash
	blockHash3 entities.BlockHash

	block1 *entities.Block
	block2 *entities.Block
	block3 *entities.Block

	mockBlockState    *mocks.MockSliceState[entities.BlockHash]
	mockInflightState *mocks.MockMapState[entities.BlockHash, bool]
	mockRPCClient     *mocks.MockClient[any]
	mockWorkerCh      chan *entities.Block
	mockMetrics       *mocks.MockRequestToNodeCounter
}

func TestBlockchainDiggerSuite(t *testing.T) {
	suite.Run(t, new(BlockchainDiggerSuite))
}

func (s *BlockchainDiggerSuite) SetupTest() {
	s.ctx = context.Background()
	s.cfg = configs.ChainConfig{
		Driver: "ethereum",
		Type:   "evm",
		Chain:  "ethereum",
		Scan: configs.ScanConfig{
			Depth: 5,
		},
	}

	s.blockHash1 = entities.BlockHash("0x123")
	s.blockHash2 = entities.BlockHash("0x456")
	s.blockHash3 = entities.BlockHash("0x789")

	s.block1 = &entities.Block{
		BlockHeader: entities.BlockHeader{
			Hash:     s.blockHash1,
			PrevHash: s.blockHash2,
			Height:   100,
		},
	}
	s.block2 = &entities.Block{
		BlockHeader: entities.BlockHeader{
			Hash:     s.blockHash2,
			PrevHash: s.blockHash3,
			Height:   99,
		},
	}
	s.block3 = &entities.Block{
		BlockHeader: entities.BlockHeader{
			Hash:     s.blockHash3,
			PrevHash: "",
			Height:   98,
		},
	}

	s.mockBlockState = mocks.NewMockSliceState[entities.BlockHash](s.T())
	s.mockInflightState = mocks.NewMockMapState[entities.BlockHash, bool](s.T())
	s.mockRPCClient = mocks.NewMockClient[any](s.T())
	s.mockWorkerCh = make(chan *entities.Block, 10)
	s.mockMetrics = mocks.NewMockRequestToNodeCounter(s.T())
}

func (s *BlockchainDiggerSuite) newDigger() *usecase.BlockchainDigger[any] {
	return usecase.NewBlockchainDigger(
		s.cfg,
		s.mockBlockState,
		s.mockInflightState,
		s.mockRPCClient,
		s.mockWorkerCh,
		usecase.BlocksJobMetrics{
			RequestToNodeCounter: s.mockMetrics,
		},
	)
}

func (s *BlockchainDiggerSuite) collectSentBlocks() []*entities.Block {
	close(s.mockWorkerCh)
	var sentBlocks []*entities.Block
	for block := range s.mockWorkerCh {
		sentBlocks = append(sentBlocks, block)
	}
	return sentBlocks
}

func (s *BlockchainDiggerSuite) TestSuccessfulProcessingOfNewBlocks() {
	// Setup expectations
	s.mockBlockState.EXPECT().Get().Return([]entities.BlockHash{})

	s.mockRPCClient.EXPECT().GetLatestBlock(s.ctx).Return(s.block1, nil)
	s.mockRPCClient.EXPECT().GetBlockByHash(s.ctx, s.blockHash2).Return(s.block2, nil)
	s.mockRPCClient.EXPECT().GetBlockByHash(s.ctx, s.blockHash3).Return(s.block3, nil)

	s.mockInflightState.EXPECT().Has(s.blockHash1).Return(false)
	s.mockInflightState.EXPECT().Has(s.blockHash2).Return(false)
	s.mockInflightState.EXPECT().Has(s.blockHash3).Return(false)
	s.mockInflightState.EXPECT().Set(s.blockHash3, true).Once().Return()
	s.mockInflightState.EXPECT().Set(s.blockHash2, true).Once().Return()
	s.mockInflightState.EXPECT().Set(s.blockHash1, true).Once().Return()

	s.mockMetrics.EXPECT().Inc(entities.Chain("ethereum")).Times(3)

	// Execute
	digger := s.newDigger()
	err := digger.Handle(s.ctx)
	s.Require().NoError(err)

	// Verify that blocks were sent to worker channel
	sentBlocks := s.collectSentBlocks()

	s.Len(sentBlocks, 3)
	s.Equal(s.block1, sentBlocks[2])
	s.Equal(s.block2, sentBlocks[1])
	s.Equal(s.block3, sentBlocks[0])
}

func (s *BlockchainDiggerSuite) TestEarlyReturnWhenBlockAlreadyInState() {
	// Setup expectations
	s.mockRPCClient.EXPECT().GetLatestBlock(s.ctx).Return(s.block1, nil)
	s.mockBlockState.EXPECT().Get().Return([]entities.BlockHash{s.blockHash1}) // Block already in state

	s.mockMetrics.EXPECT().Inc(entities.Chain("ethereum")).Times(1)

	// Execute
	digger := s.newDigger()
	err := digger.Handle(s.ctx)
	s.Require().NoError(err)

	// Verify no blocks were sent to worker channel
	sentBlocks := s.collectSentBlocks()
	s.Empty(sentBlocks)

	// Verify no RPC calls made after checking state
	s.mockRPCClient.AssertExpectations(s.T())
}

func (s *BlockchainDiggerSuite) TestEarlyBreakWhenBlockAlreadyInFlightState() {
	// Setup expectations
	s.mockBlockState.EXPECT().Get().Return([]entities.BlockHash{})

	s.mockRPCClient.EXPECT().GetLatestBlock(s.ctx).Return(s.block1, nil)
	s.mockRPCClient.EXPECT().GetBlockByHash(s.ctx, s.blockHash2).Return(s.block2, nil)

	s.mockInflightState.EXPECT().Has(s.blockHash1).Return(false) // Not in flight state
	s.mockInflightState.EXPECT().Has(s.blockHash2).Return(true)  // Block already in flight state - should break
	s.mockInflightState.EXPECT().Set(s.blockHash1, true).Once().Return()

	s.mockMetrics.EXPECT().Inc(entities.Chain("ethereum")).Times(2)

	// Execute
	digger := s.newDigger()
	err := digger.Handle(s.ctx)
	s.Require().NoError(err)

	// Verify that only one block was sent to worker channel (the first one)
	sentBlocks := s.collectSentBlocks()
	s.Len(sentBlocks, 1)
	s.Equal(s.block1, sentBlocks[0])

	// Verify inflightState was updated
	s.mockInflightState.AssertExpectations(s.T())
}

func (s *BlockchainDiggerSuite) TestScanDepthReached() {
	// Override config with limited depth
	s.cfg.Scan.Depth = 2

	// Setup expectations
	s.mockBlockState.EXPECT().Get().Return([]entities.BlockHash{})
	s.mockRPCClient.EXPECT().GetLatestBlock(s.ctx).Return(s.block1, nil)
	s.mockRPCClient.EXPECT().GetBlockByHash(s.ctx, s.blockHash2).Return(s.block2, nil)
	s.mockInflightState.EXPECT().Has(s.blockHash1).Return(false)
	s.mockInflightState.EXPECT().Has(s.blockHash2).Return(false)
	s.mockInflightState.EXPECT().Set(s.blockHash1, true).Once().Return()
	s.mockInflightState.EXPECT().Set(s.blockHash2, true).Once().Return()

	s.mockMetrics.EXPECT().Inc(entities.Chain("ethereum")).Times(2)

	// Execute
	digger := s.newDigger()
	err := digger.Handle(s.ctx)
	s.Require().NoError(err)

	// Verify that only 2 blocks were sent (due to depth limit)
	sentBlocks := s.collectSentBlocks()
	s.Len(sentBlocks, 2) // Should only get 2 blocks due to depth limit
	s.Equal(s.block1, sentBlocks[1])
	s.Equal(s.block2, sentBlocks[0])

	// Verify inflightState was updated
	s.mockInflightState.AssertExpectations(s.T())
}

func (s *BlockchainDiggerSuite) TestErrorGettingLatestBlock() {
	// Setup expectations - error on latest block
	s.mockBlockState.EXPECT().Get().Once().Return([]entities.BlockHash{})
	s.mockRPCClient.EXPECT().GetLatestBlock(s.ctx).Return(nil, errors.New("rpc error"))

	// Execute
	digger := s.newDigger()
	err := digger.Handle(s.ctx)
	s.Require().Error(err)
	s.Contains(err.Error(), "get latest block hash")

	// Verify no blocks were sent to worker channel
	sentBlocks := s.collectSentBlocks()
	s.Empty(sentBlocks)
}

func (s *BlockchainDiggerSuite) TestErrorGettingBlockByHash() {
	// Setup expectations
	s.mockRPCClient.EXPECT().GetLatestBlock(s.ctx).Return(s.block1, nil)
	s.mockBlockState.EXPECT().Get().Return([]entities.BlockHash{})
	s.mockInflightState.EXPECT().Has(s.blockHash1).Return(false)
	s.mockRPCClient.EXPECT().GetBlockByHash(s.ctx, s.blockHash2).Return(nil, errors.New("rpc error"))

	s.mockMetrics.EXPECT().Inc(entities.Chain("ethereum")).Times(1)

	// Execute
	digger := s.newDigger()
	err := digger.Handle(s.ctx)
	s.Require().Error(err)
	s.Contains(err.Error(), "get block")

	// Verify no blocks were sent to worker channel
	sentBlocks := s.collectSentBlocks()
	s.Empty(sentBlocks)
}

func (s *BlockchainDiggerSuite) TestEarlyReturnWhenLatestBlockInFlightState() {
	// Setup expectations
	s.mockBlockState.EXPECT().Get().Return([]entities.BlockHash{})

	s.mockRPCClient.EXPECT().GetLatestBlock(s.ctx).Return(s.block1, nil)

	s.mockInflightState.EXPECT().Has(s.blockHash1).Return(true) // Latest block already in flight

	s.mockMetrics.EXPECT().Inc(entities.Chain("ethereum")).Times(1)

	// Execute
	digger := s.newDigger()
	err := digger.Handle(s.ctx)
	s.Require().NoError(err)

	// Verify no blocks were sent to worker channel
	sentBlocks := s.collectSentBlocks()
	s.Empty(sentBlocks)
}

func (s *BlockchainDiggerSuite) TestBreakWhenPrevHashFoundInState() {
	// Setup expectations
	s.mockBlockState.EXPECT().Get().Return([]entities.BlockHash{s.blockHash3}) // blockHash3 is known

	s.mockRPCClient.EXPECT().GetLatestBlock(s.ctx).Return(s.block1, nil)
	s.mockRPCClient.EXPECT().GetBlockByHash(s.ctx, s.blockHash2).Return(s.block2, nil)

	s.mockInflightState.EXPECT().Has(s.blockHash1).Return(false)
	s.mockInflightState.EXPECT().Has(s.blockHash2).Return(false)
	s.mockInflightState.EXPECT().Set(s.blockHash2, true).Once().Return()
	s.mockInflightState.EXPECT().Set(s.blockHash1, true).Once().Return()

	s.mockMetrics.EXPECT().Inc(entities.Chain("ethereum")).Times(2)

	// Execute
	digger := s.newDigger()
	err := digger.Handle(s.ctx)
	s.Require().NoError(err)

	// Verify that only 2 blocks were sent (stopped when prev hash found in state)
	sentBlocks := s.collectSentBlocks()
	s.Len(sentBlocks, 2)
	s.Equal(s.block2, sentBlocks[0])
	s.Equal(s.block1, sentBlocks[1])
}
