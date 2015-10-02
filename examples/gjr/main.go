package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
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

	// Create API.
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	// Add middleware with the limiter instance.
	api.Use(limiter.NewGJRMiddleware(limiter.NewLimiter(store, rate)))

	// Set stupid app.
	api.SetApp(rest.AppSimple(func(w rest.ResponseWriter, r *rest.Request) {
		w.WriteJson(map[string]string{"message": "ok"})
	}))

	// Run server!
	fmt.Println("Server is running on 7777...")
	log.Fatal(http.ListenAndServe(":7777", api.MakeHandler()))
}
