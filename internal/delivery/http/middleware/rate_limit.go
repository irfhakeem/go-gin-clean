package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipRateLimiter struct {
	mu       sync.Mutex
	limits   map[string]*clientBucket
	rate     int           // requests allowed
	interval time.Duration // per time window
}

type clientBucket struct {
	tokens     int
	lastRefill time.Time
}

func NewRateLimiter(rate int, interval time.Duration) gin.HandlerFunc {
	limiter := &ipRateLimiter{
		limits:   make(map[string]*clientBucket),
		rate:     rate,
		interval: interval,
	}

	// Background cleaner for stale entries
	go func() {
		ticker := time.NewTicker(interval * 2)
		defer ticker.Stop()
		for range ticker.C {
			limiter.cleanup(interval * 3)
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.allow(ip) {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":  false,
				"code":    "RATE_LIMIT_EXCEEDED",
				"message": "Too many requests. Please try again later.",
			})
			return
		}
		c.Next()
	}
}

func (l *ipRateLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	bucket, exists := l.limits[ip]
	if !exists {
		l.limits[ip] = &clientBucket{
			tokens:     l.rate - 1,
			lastRefill: now,
		}
		return true
	}

	// Refill tokens if interval elapsed
	if now.Sub(bucket.lastRefill) >= l.interval {
		bucket.tokens = l.rate
		bucket.lastRefill = now
	}

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

func (l *ipRateLimiter) cleanup(maxAge time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	for ip, bucket := range l.limits {
		if now.Sub(bucket.lastRefill) > maxAge {
			delete(l.limits, ip)
		}
	}
}
