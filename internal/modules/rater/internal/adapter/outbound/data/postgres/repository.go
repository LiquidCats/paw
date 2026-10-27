package postgres

import "github.com/LiquidCats/paw/internal/modules/rater/internal/adapter/outbound/data/postgres/db"

type Repository struct {
	db *db.Queries
}

func NewRepository(db *db.Queries) *Repository {
	return &Repository{
		db: db,
	}
}
