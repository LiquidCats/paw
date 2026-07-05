package postgresql

import (
	db "github.com/LiquidCats/paw/lib/database"
	"github.com/LiquidCats/paw/services/litehsm/internal/adapter/postgresql/database"
)

type Repository struct {
	*db.QueriesTxManager[database.Queries]
}

func NewRepository(manager *db.QueriesTxManager[database.Queries]) *Repository {
	return &Repository{
		manager,
	}
}
