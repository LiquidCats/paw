package state

import (
	"sync/atomic"
)

type SliceState[T any] struct {
	capacity int

	value atomic.Value
}

func NewSliceState[T any](capacity int) *SliceState[T] {
	p := &SliceState[T]{
		capacity: capacity,
	}
	emptySlice := make([]T, 0, capacity)

	p.value.Store(&emptySlice)
	return p
}

func (s *SliceState[T]) Set(value T) {
	for {
		// Read current value
		oldValPtr := s.value.Load().(*[]T) //nolint:errcheck
		oldVal := *oldValPtr

		// Create a new slice with the added value
		newVal := make([]T, len(oldVal), len(oldVal)+1)
		copy(newVal, oldVal)
		newVal = append(newVal, value)

		// Enforce capacity limit
		if len(newVal) > s.capacity {
			newVal = newVal[1:]
		}

		// Try to atomically swap the value
		if s.value.CompareAndSwap(oldValPtr, &newVal) {
			return
		}
		// If CAS failed, retry (another goroutine modified the value)
	}
}

func (s *SliceState[T]) Get() []T {
	valPtr := s.value.Load().(*[]T) //nolint:errcheck
	val := *valPtr

	if len(val) != 0 {
		// Return a copy to prevent external modifications
		result := make([]T, len(val))
		copy(result, val)
		return result
	}

	return nil
}
