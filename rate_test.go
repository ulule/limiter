package limiter

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestRate tests Rate methods.
func TestRate(t *testing.T) {
	is := require.New(t)

	expected := map[string]Rate{
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
	}

	for k, v := range expected {
		r, err := NewRateFromFormatted(k)
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
		_, err := NewRateFromFormatted(w)
		is.Error(err)
	}

}
