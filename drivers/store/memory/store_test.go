package memory_test

import (
	"testing"
	"time"

	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/store/memory"
	"github.com/ulule/limiter/drivers/store/tests"
)

func TestMemoryStoreSequentialAccess(t *testing.T) {
	tests.TestStoreSequentialAccess(t, memory.NewStoreWithOptions(limiter.StoreOptions{
		Prefix:          "limiter:memory:sequential",
		CleanUpInterval: 30 * time.Second,
	}))
}

func TestMemoryStoreConcurrentAccess(t *testing.T) {
	tests.TestStoreConcurrentAccess(t, memory.NewStoreWithOptions(limiter.StoreOptions{
		Prefix:          "limiter:memory:concurrent",
		CleanUpInterval: 1 * time.Nanosecond,
	}))
}
