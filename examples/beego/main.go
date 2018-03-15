

import (
	"github.com/astaxie/beego"
	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/store/memory"
	"github.com/astaxie/beego/context"
	"net/http"
	"strconv"
)

var rateLimiter *limiter.Limiter

func main() {
	rate, err := limiter.NewRateFromFormatted("2-H")
	if err != nil{
		panic(err)
	}
	store := memory.NewStore()

	rateLimiter = limiter.New(store, rate)

	//More on Beego filters here https://beego.me/docs/mvc/controller/filter.md
	beego.InsertFilter("*", beego.BeforeRouter, rateLimit)
	beego.Run()
}

func rateLimit(ctx *context.Context) {
	r := ctx.Request

	context, err := rateLimiter.Get(r.Context(), limiter.GetIPKey(r, false))
	if err != nil {
		ctx.Abort(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	h := ctx.ResponseWriter.Header()
	h.Add("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
	h.Add("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
	h.Add("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

	if context.Reached {
		//This will cause a panic on the logs. To avoid this, add the error string to 
		//Beego.ErrorMaps as discribed here https://beego.me/docs/mvc/controller/errors.md
		ctx.Abort(http.StatusTooManyRequests, "Too Many Requests")
		return
	}

}
