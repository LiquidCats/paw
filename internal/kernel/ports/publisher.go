package ports

import "context"

type TopicAwarePublisher[T any] interface {
	PublishTo(ctx context.Context, topic string, data T) error
}

type TopicPublisher[T any] interface {
	Publish(ctx context.Context, data T) error
}
