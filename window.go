package erl

import (
	"fmt"
	"time"
)

// Window represents a time window for rate limit tracking.
type Window int

const (
	// PerMinute tracks requests in one-minute buckets.
	PerMinute Window = iota
	// PerHour tracks requests in one-hour buckets.
	PerHour
	// PerDay tracks requests in 24-hour (calendar day, UTC) buckets.
	PerDay
	// PerMonth tracks requests in 30-day (calendar month, UTC) buckets.
	PerMonth
)

// Duration returns the duration of the window.
func (w Window) Duration() time.Duration {
	switch w {
	case PerMinute:
		return time.Minute
	case PerHour:
		return time.Hour
	case PerDay:
		return 24 * time.Hour
	case PerMonth:
		return 30 * 24 * time.Hour
	default:
		return time.Hour
	}
}

// BucketKey returns a time-bucket suffix for the current moment.
// This is used to partition counters by window period.
func (w Window) BucketKey(t time.Time) string {
	t = t.UTC()
	switch w {
	case PerMinute:
		return t.Format("2006-01-02T15:04")
	case PerHour:
		return t.Format("2006-01-02T15")
	case PerDay:
		return t.Format("2006-01-02")
	case PerMonth:
		return t.Format("2006-01")
	default:
		return t.Format("2006-01-02T15")
	}
}

// BucketStart returns the start time of the current bucket.
func (w Window) BucketStart(t time.Time) time.Time {
	t = t.UTC()
	switch w {
	case PerMinute:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)
	case PerHour:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.UTC)
	case PerDay:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	case PerMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	default:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.UTC)
	}
}

func (w Window) String() string {
	switch w {
	case PerMinute:
		return "PerMinute"
	case PerHour:
		return "PerHour"
	case PerDay:
		return "PerDay"
	case PerMonth:
		return "PerMonth"
	default:
		return fmt.Sprintf("Window(%d)", int(w))
	}
}
