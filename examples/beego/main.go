package main

import (
	"github.com/astaxie/beego"
	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/store/memory"
	"github.com/astaxie/beego/context"
	"net/http"
	"strconv"
	"strings"
	"github.com/astaxie/beego/logs"
)

type rateLimiter struct {
	GeneralLimiter *limiter.Limiter
	LoginLimiter   *limiter.Limiter
}

func main() {
	rates := new(rateLimiter)
	store := memory.NewStore()
	
	rate, err := limiter.NewRateFromFormatted("2-S")
	PanicOnError(err)
	rates.GeneralLimiter = limiter.New(store, rate)

	loginRate, err := limiter.NewRateFromFormatted("2-M")
	PanicOnError(err)
	rates.LoginLimiter = limiter.New(store, loginRate)

	//More on Beego filters here https://beego.me/docs/mvc/controller/filter.md
	beego.InsertFilter("/*", beego.BeforeRouter, func(c *context.Context) {
		rateLimit(rates, c)
	}, true)

	beego.ErrorHandler("429", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Too Many Requests"))
		return
	})
	beego.Run()
}

func rateLimit(l *rateLimiter, ctx *context.Context) {
	var (
		context limiter.Context
		err     error
		r       = ctx.Request
		ip      = limiter.GetIPKey(r, false)
	)

	if strings.HasPrefix(ctx.Input.URL(), "/login") {
		context, err = l.LoginLimiter.Get(r.Context(), ip)
	} else {
		context, err = l.GeneralLimiter.Get(r.Context(), ip)
	}
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, err.Error())
		return
	}

	h := ctx.ResponseWriter.Header()
	h.Add("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
	h.Add("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
	h.Add("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

	if context.Reached {
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
