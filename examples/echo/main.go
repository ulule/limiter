package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

var (
	ipRateLimiter *limiter.Limiter
	store         limiter.Store
)

func main() {
	e := echo.New()
	e.GET("/hello", hello, IPRateLimit()) // 3. Use middleware
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", 5555)))
}

func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func IPRateLimit() echo.MiddlewareFunc {
	// 1. Configure
	rate := limiter.Rate{
		Period: 2 * time.Second,
		Limit:  1,
	}
	store = memory.NewStore()
	ipRateLimiter = limiter.New(store, rate)

	// 2. Return middleware handler
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			ip := c.RealIP()
			limiterCtx, err := ipRateLimiter.Get(c.Request().Context(), ip)
			if err != nil {
				log.Printf("IPRateLimit - ipRateLimiter.Get - err: %v, %s on %s", err, ip, c.Request().URL)
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"success": false,
					"message": err,
				})
			}

			h := c.Response().Header()
			h.Set("X-RateLimit-Limit", strconv.FormatInt(limiterCtx.Limit, 10))
			h.Set("X-RateLimit-Remaining", strconv.FormatInt(limiterCtx.Remaining, 10))
			h.Set("X-RateLimit-Reset", strconv.FormatInt(limiterCtx.Reset, 10))

			if limiterCtx.Reached {
				log.Printf("Too Many Requests from %s on %s", ip, c.Request().URL)
				return c.JSON(http.StatusTooManyRequests, echo.Map{
					"success": false,
					"message": "Too Many Requests on " + c.Request().URL.String(),
				})
			}

			// log.Printf("%s request continue", c.RealIP())
			return next(c)
		}
	}
}
