package main

/*
More comprehensive example:
https://gist.github.com/gadelkareem/5a087bfda1f673241d0ac65759156cfd
*/
import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/logs"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"net/http"
	"strconv"
	"strings"
)

type rateLimiter struct {
	generalLimiter *limiter.Limiter
	loginLimiter   *limiter.Limiter
}

func main() {
	r := &rateLimiter{}

	rate, err := limiter.NewRateFromFormatted("2-S")
	PanicOnError(err)
	r.generalLimiter = limiter.New(memory.NewStore(), rate)

	loginRate, err := limiter.NewRateFromFormatted("2-M")
	PanicOnError(err)
	r.loginLimiter = limiter.New(memory.NewStore(), loginRate)

	//More on Beego filters here https://beego.me/docs/mvc/controller/filter.md
	beego.InsertFilter("/*", beego.BeforeRouter, func(c *context.Context) {
		rateLimit(r, c)
	}, true)

	//refer to https://beego.me/docs/mvc/controller/errors.md for error handling
	beego.ErrorHandler("429", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Too Many Requests"))
		return
	})
	beego.Run()
}

func rateLimit(r *rateLimiter, ctx *context.Context) {
	var (
		limiterCtx limiter.Context
		err        error
		req        = ctx.Request
	)

	ip := limiter.GetIP(req, false)

	if strings.HasPrefix(ctx.Input.URL(), "/login") {
		limiterCtx, err = r.loginLimiter.Get(req.Context(), ip.String())
	} else {
		limiterCtx, err = r.generalLimiter.Get(req.Context(), ip.String())
	}
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}

	h := ctx.ResponseWriter.Header()
	h.Add("X-RateLimit-Limit", strconv.FormatInt(limiterCtx.Limit, 10))
	h.Add("X-RateLimit-Remaining", strconv.FormatInt(limiterCtx.Remaining, 10))
	h.Add("X-RateLimit-Reset", strconv.FormatInt(limiterCtx.Reset, 10))

	if limiterCtx.Reached {
		logs.Debug("Too Many Requests from %s on %s", ip, ctx.Input.URL())
		//refer to https://beego.me/docs/mvc/controller/errors.md for error handling
		ctx.Abort(http.StatusTooManyRequests, "429")
		return
	}

}

func PanicOnError(e error) {
	if e != nil {
		panic(e)
	}
}
