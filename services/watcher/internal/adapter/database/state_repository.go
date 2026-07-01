package database

import (
	"context"
	"fmt"

	"github.com/LiquidCats/paw/services/watcher/internal/app/domain/entities"
	"github.com/bytedance/sonic"
	"github.com/rotisserie/eris"
)

const blockStateKey = "blocks-state"

func (r *Repository) SetBlockState(
	ctx context.Context,
	chain entities.Chain,
	blocks []entities.BlockHash,
) error {
	key := fmt.Sprintf("%s:%s", blockStateKey, chain)
	value, err := sonic.Marshal(blocks)
	if err != nil {
		return eris.Wrap(err, "marshal blocks")
	}

	err = r.queries.SetState(ctx, SetStateParams{
		Key:   key,
		Value: value,
	})
	if err != nil {
		return eris.Wrap(err, "set block state")
	}

	return nil
}
