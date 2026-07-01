package database

type Repository struct {
	queries Queries
}

func NewRepository(db DBTX) *Repository {
	q := New(db)
	return &Repository{
		queries: *q,
	}
}
