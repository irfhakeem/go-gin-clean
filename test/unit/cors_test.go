package unit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-gin-clean/internal/delivery/http/middleware"

	"github.com/gin-gonic/gin"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	allowed := []string{"http://localhost:3000", "https://app.example.com"}

	t.Run("AC-010: Allowed Origin Receives Specific Header and Credentials", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.CORS(allowed...))
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "https://app.example.com")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Header().Get("Access-Control-Allow-Origin") != "https://app.example.com" {
			t.Fatalf("expected origin 'https://app.example.com', got %q", w.Header().Get("Access-Control-Allow-Origin"))
		}
		if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Fatalf("expected credentials true, got %q", w.Header().Get("Access-Control-Allow-Credentials"))
		}
	})

	t.Run("AC-010: Non-Allowed Origin Does Not Receive Allow Header", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.CORS(allowed...))
		router.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "https://attacker.evil.com")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatalf("expected empty allow-origin for untrusted origin, got %q", w.Header().Get("Access-Control-Allow-Origin"))
		}
	})
}
