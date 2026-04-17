package middleware

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"github.com/Shravanthi20/InDel/backend/internal/database"
	"github.com/gin-gonic/gin"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func generateCacheKey(c *gin.Context) string {
	url := c.Request.URL.RequestURI()
	hash := md5.Sum([]byte(url))
	return "cache:" + hex.EncodeToString(hash[:])
}

// RedisCache caches the response for a given TTL using the request URI.
func RedisCache(ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		if database.Rdb == nil {
			// Redis not initialized, bypass
			c.Next()
			return
		}

		cacheKey := generateCacheKey(c)

		// Try loading from cache
		cachedData, err := database.Rdb.Get(context.Background(), cacheKey).Result()
		if err == nil && len(cachedData) > 0 {
			// Cache hit
			c.Header("X-Cache", "HIT")
			c.Data(http.StatusOK, "application/json", []byte(cachedData))
			c.Abort()
			return
		}

		// Cache miss
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		statusCode := c.Writer.Status()
		if statusCode >= 200 && statusCode < 300 {
			// Save to cache asynchronously or synchronously
			err := database.Rdb.Set(context.Background(), cacheKey, blw.body.Bytes(), ttl).Err()
			if err != nil {
				log.Printf("Failed to set cache for %s: %v", cacheKey, err)
			}
			c.Header("X-Cache", "MISS")
		}
	}
}

// RedisInvalidate dynamically clears wildcards
func RedisInvalidate(pattern string) {
	if database.Rdb == nil {
		return
	}
	ctx := context.Background()
	keys, err := database.Rdb.Keys(ctx, pattern).Result()
	if err == nil && len(keys) > 0 {
		database.Rdb.Del(ctx, keys...)
	}
}
