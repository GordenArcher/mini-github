package middleware

import (
	"fmt"
	"time"

	"github.com/GordenArcher/mini-github/internal/helper/responses"
	"github.com/GordenArcher/mini-github/internal/redis"
	"github.com/gin-gonic/gin"
	limiterlib "github.com/ulule/limiter/v3"
	limiterredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

func RateLimitMiddleware() gin.HandlerFunc {
	store, err := limiterredis.NewStore(redis.Client)
	if err != nil {
		panic(err)
	}

	rate := limiterlib.Rate{Period: time.Minute, Limit: 10}
	limiter := limiterlib.New(store, rate)

	return func(c *gin.Context) {
		key := c.ClientIP()
		context, err := limiter.Get(c, key)
		if err != nil {
			responses.JSONError(c, 429, "rate limit error")
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", context.Limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", context.Remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", int(context.Reset)))

		if context.Reached {
			responses.JSONError(c, 429, "rate limit exceeded")
			return
		}

		c.Next()
	}
}
