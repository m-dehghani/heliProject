package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

var idempotencyCache = cache.New(5*time.Minute, 10*time.Minute)

func Idempotency(c *gin.Context) {
	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing idempotency key"})
		c.Abort()
		return
	}

	if _, found := idempotencyCache.Get(idempotencyKey); found {
		c.JSON(http.StatusConflict, gin.H{"error": "duplicate request"})
		c.Abort()
		return
	}

	idempotencyCache.Set(idempotencyKey, true, cache.DefaultExpiration)
	c.Next()
}
