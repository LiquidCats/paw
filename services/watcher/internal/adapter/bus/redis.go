package bus

import (
	"context"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
	"github.com/rotisserie/eris"
)

type RedisPublisher[T any] struct {
	client redis.UniversalClient
}

func NewRedisPublisher[T any](client redis.UniversalClient) *RedisPublisher[T] {
	return &RedisPublisher[T]{
		client: client,
	}
}

func (p *RedisPublisher[T]) pub(ctx context.Context, channel string, data any) error {
	dataBytes, err := sonic.ConfigFastest.Marshal(data)
	if err != nil {
		return eris.Wrap(err, "marshal data")
	}

	if err = p.client.Publish(ctx, channel, dataBytes).Err(); err != nil {
		return eris.Wrap(err, "publish message")
	}

	return nil
}

func (p *RedisPublisher[T]) PublishTo(ctx context.Context, topic string, data T) error {
	if err := p.pub(ctx, topic, data); err != nil {
		return eris.Wrap(err, "publish transaction")
	}

	return nil
}
