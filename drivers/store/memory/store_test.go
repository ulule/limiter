package memory_test

import (
	"testing"
	"time"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"github.com/ulule/limiter/v3/drivers/store/tests"
)

func TestMemoryStoreSequentialAccess(t *testing.T) {
	tests.TestStoreSequentialAccess(t, memory.NewStoreWithOptions(limiter.StoreOptions{
		Prefix:          "limiter:memory:sequential-test",
		CleanUpInterval: 30 * time.Second,
	}))
}

func TestMemoryStoreConcurrentAccess(t *testing.T) {
	tests.TestStoreConcurrentAccess(t, memory.NewStoreWithOptions(limiter.StoreOptions{
		Prefix:          "limiter:memory:concurrent-test",
		CleanUpInterval: 1 * time.Nanosecond,
	}))
}

func BenchmarkMemoryStoreSequentialAccess(b *testing.B) {
	tests.BenchmarkStoreSequentialAccess(b, memory.NewStoreWithOptions(limiter.StoreOptions{
		Prefix:          "limiter:memory:sequential-benchmark",
		CleanUpInterval: 1 * time.Hour,
	}))
}

func BenchmarkMemoryStoreConcurrentAccess(b *testing.B) {
	tests.BenchmarkStoreConcurrentAccess(b, memory.NewStoreWithOptions(limiter.StoreOptions{
		Prefix:          "limiter:memory:concurrent-benchmark",
		CleanUpInterval: 1 * time.Hour,
	}))
}
