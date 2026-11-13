package unit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-gin-clean/internal/delivery/http/middleware"

	"github.com/gin-gonic/gin"
)

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("AC-007: Throttles Burst Requests with 429", func(t *testing.T) {
		router := gin.New()
		// Allow 3 requests per 500ms
		limiter := middleware.NewRateLimiter(3, 500*time.Millisecond)
		router.Use(limiter)
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		for i := 1; i <= 3; i++ {
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("request %d expected 200 OK, got %d", i, w.Code)
			}
		}

		// 4th request must be throttled with 429
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusTooManyRequests {
			t.Fatalf("request 4 expected 429 Too Many Requests, got %d", w.Code)
		}
	})
}
