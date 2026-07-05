package postgresql

import (
	"context"
	"fmt"
	"time"

	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/postgresql/database"
	"github.com/LiquidCats/paw/services/litehsm/internal/app/domain/entities"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *Repository) CreateKey(ctx context.Context, entry entities.KeyEntry) error {
	params := database.CreateKeyParams{
		SeedFingerprint: entry.SeedFingerprint,
		Alias:           entry.Alias,
		Curve:           string(entry.Curve),
		Algorithm:       string(entry.Algorithm),
		DerivationPath:  entry.DerivationPath.String(),
		Status:          string(entry.Status),
		ExpiresAt:       pgtype.Timestamp{},
	}

	if entry.ExpiresAt != nil {
		params.ExpiresAt.Time = *entry.ExpiresAt
		params.ExpiresAt.Valid = true
	}

	_, err := r.GetQueries(ctx).CreateKey(ctx, params)
	if err != nil {
		return fmt.Errorf("struct=Repository, method=CreateKey, call=queries.CreateKey: %w", err)
	}

	return nil
}

func (r *Repository) GetAllKeys(ctx context.Context) ([]entities.KeyEntry, error) {
	keys, err := r.GetQueries(ctx).GetAllKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("struct=Repository, method=GetKey, call=queries.GetAllKeys: %w", err)
	}

	entries := make([]entities.KeyEntry, len(keys))

	for i, key := range keys {
		derPath, err := entities.ParseDerivationPath(key.DerivationPath)
		if err != nil {
			return nil, fmt.Errorf("struct=Repository, method=GetAllKeys, call=ParseDerivationPath: %w", err)
		}

		entry := entities.KeyEntry{
			SeedFingerprint: key.SeedFingerprint,
			Alias:           key.Alias,
			Curve:           entities.CurveType(key.Curve),
			Algorithm:       entities.AlgorithmType(key.Algorithm),
			DerivationPath:  derPath,
			Status:          entities.KeyStatus(key.Status),
		}
		if key.ExpiresAt.Valid {
			entry.ExpiresAt = new(key.ExpiresAt.Time)
		}

		copy(entry.KeyID[:], key.KeyID.Bytes[:])

		entries[i] = entry
	}

	return entries, nil
}

func (r *Repository) GetKey(ctx context.Context, keyID entities.KeyID) (*entities.KeyEntry, error) {
	key, err := r.GetQueries(ctx).GetKey(ctx, pgtype.UUID{
		Bytes: keyID,
		Valid: true,
	})
	if err != nil {
		return nil, fmt.Errorf("get key query: %w", err)
	}

	derPath, err := entities.ParseDerivationPath(key.DerivationPath)
	if err != nil {
		return nil, fmt.Errorf("struct=Repository, method=GetKey, call=ParseDerivationPath: %w", err)
	}

	entry := entities.KeyEntry{
		KeyID:           keyID,
		SeedFingerprint: key.SeedFingerprint,
		Alias:           key.Alias,
		Curve:           entities.CurveType(key.Curve),
		Algorithm:       entities.AlgorithmType(key.Algorithm),
		DerivationPath:  derPath,
		Status:          entities.KeyStatus(key.Status),
	}

	if key.ExpiresAt.Valid {
		entry.ExpiresAt = new(key.ExpiresAt.Time)
	}

	return &entry, nil
}

func (r *Repository) SetExpiration(ctx context.Context, keyID entities.KeyID, expiresAt *time.Time) error {
	err := r.GetQueries(ctx).SetExpiration(ctx, database.SetExpirationParams{
		KeyID:     pgtype.UUID{Bytes: keyID, Valid: true},
		ExpiresAt: pgtype.Timestamp{Time: *expiresAt, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("struct=Repository, method=SetExpiration, call=queries.SetExpiration: %w", err)
	}

	return nil
}

func (r *Repository) SetStatus(ctx context.Context, keyID entities.KeyID, status entities.KeyStatus) error {
	err := r.GetQueries(ctx).SetStatus(ctx, database.SetStatusParams{
		KeyID:  pgtype.UUID{Bytes: keyID, Valid: true},
		Status: string(status),
	})
	if err != nil {
		return fmt.Errorf("struct=Repository, method=SetExpiration, call=queries.SetStatus: %w", err)
	}

	return nil
}
