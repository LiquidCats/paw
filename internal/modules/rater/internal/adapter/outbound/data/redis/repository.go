package redis

import (
	goredis "github.com/redis/go-redis/v9"
)

type Repository struct {
	client goredis.UniversalClient
}

func NewRepository(client goredis.UniversalClient) *Repository {
	return &Repository{client: client}
}
