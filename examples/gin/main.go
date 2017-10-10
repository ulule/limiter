package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/ulule/limiter"
	mgin "github.com/ulule/limiter/drivers/middleware/gin"
	sredis "github.com/ulule/limiter/drivers/store/redis"
)

func main() {

	// Define a limit rate to 4 requests per hour.
	rate, err := limiter.NewRateFromFormatted("4-H")
	if err != nil {
		log.Fatal(err)
		return
	}

	// Create a redis client.
	option, err := redis.ParseURL("redis://localhost:6379/0")
	if err != nil {
		log.Fatal(err)
		return
	}
	client := redis.NewClient(option)

	// Create a store with the redis client.
	store, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "limiter_gin_example",
		MaxRetry: 3,
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	// Create a new middleware with the limiter instance.
	middleware := mgin.NewMiddleware(limiter.New(store, rate))

	// Launch a simple server.
	router := gin.Default()
	router.Use(middleware)
	router.GET("/", index)
	log.Fatal(router.Run(":7777"))
}

func index(c *gin.Context) {
	type message struct {
		Message string `json:"message"`
	}
	resp := message{Message: "ok"}
	c.JSON(http.StatusOK, resp)
}
