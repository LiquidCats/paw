package runner

import "context"

type Job interface {
	Handle(ctx context.Context) error
}
