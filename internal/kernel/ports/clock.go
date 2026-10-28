package ports

import (
	"time"

	"github.com/LiquidCats/paw/internal/kernel/domain/values"
)

type Clock interface {
	Now() time.Time
	// RoundToNearest rounds a time to the nearest 5-minute interval.
	RoundToNearest(t time.Time, b values.ClockBucket) time.Time
}
