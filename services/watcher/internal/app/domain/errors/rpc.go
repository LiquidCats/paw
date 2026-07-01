package errors // nolint:revive

import (
	"fmt"
)

type RPCAdapterError struct {
	Message string
	Params  []any
}

func (e RPCAdapterError) Error() string {
	return fmt.Sprintf("rpc error: message=%q params=%v", e.Message, e.Params)
}

func NewPossibleReorgError[T comparable](p T) error {
	return RPCAdapterError{
		Message: "possible reorg on block",
		Params:  []any{p},
	}
}
