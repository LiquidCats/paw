package runner

import "context"

type ChanWrite[T any] chan<- T
type ChanRead[T any] <-chan T

type Handler[T any] interface {
	Handle(ctx context.Context, value T) error
}
