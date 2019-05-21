package limiter_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ulule/limiter"
)

// TestRate tests Rate methods.
func TestRate(t *testing.T) {
	is := require.New(t)

	expected := map[string]limiter.Rate{
		"10-S": {
			Formatted: "10-S",
			Period:    1 * time.Second,
			Limit:     int64(10),
		},
		"356-M": {
			Formatted: "356-M",
			Period:    1 * time.Minute,
			Limit:     int64(356),
		},
		"3-H": {
			Formatted: "3-H",
			Period:    1 * time.Hour,
			Limit:     int64(3),
		},
		"2000-D": {
			Formatted: "2000-D",
			Period:    24 * time.Hour,
			Limit:     int64(2000),
		},
	}

	for k, v := range expected {
		r, err := limiter.NewRateFromFormatted(k)
		is.NoError(err)
		is.True(reflect.DeepEqual(v, r))
	}

	wrongs := []string{
		"10 S",
		"10:S",
		"AZERTY",
		"na wak",
		"H-10",
	}

	for _, w := range wrongs {
		_, err := limiter.NewRateFromFormatted(w)
		is.Error(err)
	}

}
