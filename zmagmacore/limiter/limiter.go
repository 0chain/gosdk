// DEPRECATED: This package is deprecated and will be removed in a future release.
package limiter

import (
	"net/http"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
)

// rateLimiter represents custom wrapper above limiter.Limiter.
type rateLimit struct {
	Limiter           *limiter.Limiter
	RateLimit         bool
	RequestsPerSecond float64
}

// userRateLimit represents application level limiter.
var userRateLimit *rateLimit

func (rl *rateLimit) init() {
	if rl.RequestsPerSecond == 0 {
		rl.RateLimit = false
		return
	}
	rl.RateLimit = true
	rl.Limiter = tollbooth.NewLimiter(rl.RequestsPerSecond, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour}).
		SetIPLookups([]string{"RemoteAddr", "X-Forwarded-For", "X-Real-IP"}).
		SetMethods([]string{"GET", "POST", "PUT", "DELETE"})
}

// ConfigRateLimits configures rate limits used in app.
//
// Should be called only once while application starting process.
func ConfigRateLimits(limit float64) {
	userRateLimit = &rateLimit{RequestsPerSecond: limit}
	userRateLimit.init()
}

// UserRateLimit is a middleware that performs rate-limiting given request handler function.
func UserRateLimit(handler http.HandlerFunc) http.HandlerFunc {
	if !userRateLimit.RateLimit {
		return handler
	}
	return func(writer http.ResponseWriter, request *http.Request) {
		tollbooth.LimitFuncHandler(userRateLimit.Limiter, handler).ServeHTTP(writer, request)
	}
}
