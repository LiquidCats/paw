package usecase

import (
	"context"

	"github.com/LiquidCats/paw/services/watcher/configs"
	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
	"github.com/LiquidCats/paw/services/watcher/internal/app/port/database"
	"github.com/LiquidCats/paw/services/watcher/internal/app/port/state"
	"github.com/rotisserie/eris"
)

type BlocksPersister struct {
	cfg   configs.ChainConfig
	db    database.StateDB
	state state.SliceState[entities.BlockHash]
}

func NewBlocksPersister(
	cfg configs.ChainConfig,
	db database.StateDB,
	state state.SliceState[entities.BlockHash],
) *BlocksPersister {
	return &BlocksPersister{
		cfg:   cfg,
		db:    db,
		state: state,
	}
}

func (uc *BlocksPersister) Handle(ctx context.Context) error {
	blockHashes := uc.state.Get()
	err := uc.db.SetBlockState(ctx, uc.cfg.Chain, blockHashes)
	if err != nil {
		return eris.Wrap(err, "set blocks state")
	}

	return nil
}
