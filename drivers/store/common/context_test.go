package common_test

import (
	"testing"
	"time"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/common"
)

func TestGetContextFromStateReset(t *testing.T) {
	tests := []struct {
		name       string
		now        time.Time
		expiration time.Time
		wantReset  int64
	}{
		{
			name:       "exact second",
			now:        time.Unix(100, 0),
			expiration: time.Unix(101, 0),
			wantReset:  101,
		},
		{
			name:       "one nanosecond past second",
			now:        time.Unix(100, 0),
			expiration: time.Unix(101, 1),
			wantReset:  102,
		},
		{
			name:       "fractional second",
			now:        time.Unix(100, 0),
			expiration: time.Unix(101, 500000000),
			wantReset:  102,
		},
		{
			name:       "last nanosecond of second",
			now:        time.Unix(100, 0),
			expiration: time.Unix(101, 999999999),
			wantReset:  102,
		},
		{
			name:       "unexpired limit within current second",
			now:        time.Unix(101, 100000000),
			expiration: time.Unix(101, 200000000),
			wantReset:  102,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate := limiter.Rate{Limit: 2, Period: time.Second}
			ctx := common.GetContextFromState(tt.now, rate, tt.expiration, 3)
			if ctx.Reset != tt.wantReset {
				t.Errorf("Reset = %d, want %d", ctx.Reset, tt.wantReset)
			}
			reset := time.Unix(ctx.Reset, 0)
			if reset.Before(tt.expiration) {
				t.Errorf("reset time %s precedes expiration %s", reset, tt.expiration)
			}
			if reset.Sub(tt.expiration) >= time.Second {
				t.Errorf("reset time %s adds at least a full second to expiration %s", reset, tt.expiration)
			}
			if !reset.After(tt.now) {
				t.Errorf("unexpired limit has non-positive wait: reset %s, now %s", reset, tt.now)
			}
			if ctx.Limit != 2 || ctx.Remaining != 0 || !ctx.Reached {
				t.Errorf("unexpected quota state: %+v", ctx)
			}
		})
	}
}
