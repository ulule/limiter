package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/garyburd/redigo/redis"
	"github.com/ulule/limiter"
)

func main() {
	// 4 reqs/hour
	rate, err := limiter.NewRateFromFormatted("4-H")
	if err != nil {
		panic(err)
	}

	// Create a Redis pool.
	pool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", ":6379")
		if err != nil {
			return nil, err
		}
		return c, err
	}, 100)

	// Create a store with the pool.
	store, err := limiter.NewRedisStore(pool, "limitergjrexample")
	if err != nil {
		panic(err)
	}

	mw := limiter.NewHTTPMiddleware(limiter.NewLimiter(store, rate))
	http.Handle("/", mw.Handler(http.HandlerFunc(index)))

	fmt.Println("Server is runnnig on port 7777...")
	log.Fatal(http.ListenAndServe(":7777", nil))

}

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "ok"}`))
}
