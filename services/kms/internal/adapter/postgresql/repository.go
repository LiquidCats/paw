package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/postgresql/database"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/ports"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	tx      ports.TxManager
	queries *database.Queries
}

type queriesKey struct{}

func NewRepository(queries *database.Queries, manager ports.TxManager) *Repository {
	return &Repository{
		tx:      manager,
		queries: queries,
	}
}

func (r *Repository) WithinTx(ctx context.Context, callback ports.TxCallback) error {
	return r.tx.Transactional(ctx, func(txCtx context.Context) error {
		tx, ok := txCtx.Value(dbTxKey{}).(pgx.Tx)
		if !ok {
			return errors.New("struct=Repository, method=WithinTx, call=txCtx.Value: tx not found in context")
		}

		queriesCtx := context.WithValue(ctx, queriesKey{}, r.queries.WithTx(tx))

		if err := callback(queriesCtx); err != nil {
			return fmt.Errorf("struct=Repository, method=WithinTx, call=callback: %w", err)
		}

		return nil
	})
}

func (r *Repository) getQueries(ctx context.Context) *database.Queries {
	queriesTx, ok := ctx.Value(queriesKey{}).(*database.Queries)
	if !ok {
		return r.queries
	}

	return queriesTx
}
