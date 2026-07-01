package bus

import "context"

type Publisher[T any] interface {
	PublishTo(ctx context.Context, topic string, data T) error
}
