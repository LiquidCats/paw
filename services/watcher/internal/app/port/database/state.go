package database

import (
	"context"

	"github.com/LiquidCats/paw/watcher/internal/app/domain/entities"
)

type StateDB interface {
	SetBlockState(
		ctx context.Context,
		chain entities.Chain,
		blocks []entities.BlockHash,
	) error
}
