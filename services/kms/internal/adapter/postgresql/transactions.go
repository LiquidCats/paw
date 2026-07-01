package postgresql

import (
	"context"
	"fmt"

	"github.com/LiquidCats/paw/services/litehsm/internal/app/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxManager struct {
	pool *pgxpool.Pool
}

type dbTxKey struct{}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) Transactional(ctx context.Context, callback ports.TxCallback) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	txCtx := context.WithValue(ctx, dbTxKey{}, tx)

	if err = callback(txCtx); err != nil {
		return fmt.Errorf("execute transaction: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	_ = txCtx

	return nil
}

var _ ports.TxManager = (*TxManager)(nil)
