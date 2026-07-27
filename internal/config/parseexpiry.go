// config/parseexpiry.go
package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseExpiry parses a duration string, extending Go's time.ParseDuration with
// support for a "d" (days) suffix — e.g. "7d" — since Go's stdlib only understands
// ns/us/ms/s/m/h natively, and JWT_REFRESH_EXPIRY is expressed in days for readability.
func ParseExpiry(s string) (time.Duration, error) {
	// TODO 1: check if s ends with "d" (hint: strings.HasSuffix(s, "d")).
	// If it doesn't, this isn't a days-format string — just delegate straight
	// to time.ParseDuration(s) and return its result directly (handles "1h", "30m", etc).

	if !strings.HasSuffix(s, "d") {
		result, err := time.ParseDuration(s)
		if err != nil {
			return 0, err
		}
		return result, nil
	}

	// TODO 2: if it DOES end with "d":
	//   a) strip the trailing "d" (hint: strings.TrimSuffix(s, "d"))
	str := strings.TrimSuffix(s, "d")
	//   b) parse what's left as an integer (hint: strconv.Atoi) — this is the number of days
	day, err := strconv.Atoi(str)

	if err != nil {
		return 0, fmt.Errorf("error occured: %w", err)
	}

	if day <= 0 {
		return 0, fmt.Errorf("invalid refresh expiry")
	}
	duration := time.Duration(day) * 24 * time.Hour
	//   c) if that parse fails, return a wrapped error via fmt.Errorf("...: %w", err) —
	//      don't swallow it, the caller needs to know "7dd" or "xd" is malformed input
	//   d) convert days to a time.Duration: time.Duration(days) * 24 * time.Hour
	//   e) return that duration, nil

	return duration, nil
}
