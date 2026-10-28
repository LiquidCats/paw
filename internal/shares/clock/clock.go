package clock

import (
	"time"

	"github.com/LiquidCats/paw/internal/kernel/domain/values"
	"github.com/LiquidCats/paw/internal/kernel/ports"
)

var _ ports.Clock = (*Clock)(nil)

type Clock struct {
}

func NewClock() *Clock {
	return &Clock{}
}

// Now implements [ports.Clock].
func (c *Clock) Now() time.Time {
	return time.Now()
}

// RoundToNearest implements [ports.Clock].
func (c *Clock) RoundToNearest(t time.Time, b values.ClockBucket) time.Time {
	// Get the minute component
	minute := t.Minute()

	// Calculate the remainder when divided by 5
	remainder := minute % int(b)

	// Determine whether to round up or down
	var roundedMinute int
	if remainder < 3 { //nolint:mnd
		// Round down
		roundedMinute = minute - remainder
	} else {
		// Round up
		roundedMinute = minute + (5 - remainder) //nolint:mnd
	}

	// Create a new time with the rounded minutes
	rounded := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), roundedMinute, 0, 0, t.Location())

	// Handle the case where rounding up goes to the next hour
	if roundedMinute >= 60 { //nolint:mnd
		rounded = rounded.Add(-time.Duration(roundedMinute-60) * time.Minute) //nolint:mnd
	}

	return rounded
}
