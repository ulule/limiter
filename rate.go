package limiter

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Rate is the rate.
type Rate struct {
	Formatted string
	Period    time.Duration
	Limit     int64
}

// NewRateFromFormatted returns the rate from the formatted version.
func NewRateFromFormatted(formatted string) (Rate, error) {
	rate := Rate{}

	values := strings.Split(formatted, "-")
	if len(values) != 2 {
		return rate, fmt.Errorf("Incorrect format '%s'", formatted)
	}

	periods := map[string]time.Duration{
		"S": time.Second, // Second
		"M": time.Minute, // Minute
		"H": time.Hour,   // Hour
	}

	limit, period := values[0], strings.ToUpper(values[1])

	duration, ok := periods[period]

	if !ok {
		return rate, fmt.Errorf("Incorrect period '%s'", period)
	}

	var (
		p time.Duration
		l int
	)

	p = 1 * duration

	l, err := strconv.Atoi(limit)
	if err != nil {
		return rate, fmt.Errorf("Incorrect limit '%s'", limit)
	}

	return Rate{
		Formatted: formatted,
		Period:    p,
		Limit:     int64(l),
	}, nil
}
