package assets

import "time"

func FutureTime() time.Time {
	return time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)
}

func PastTimeWithHighYearDay() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year()-1, time.December, 31, 0, 0, 0, 0, time.UTC)
}
