package custom_middlewares

import (
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(rate.Every(1*time.Minute), 2)

func RateLimitMiddleware(api huma.API) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// Set a custom header on the response.
		if !limiter.Allow() {
			_ = huma.WriteErr(api, ctx, http.StatusTooManyRequests, "Too many requests")
			return
		}

		// Call the next middleware in the chain. This eventually calls the
		// operation handler as well.
		next(ctx)
	}
}
