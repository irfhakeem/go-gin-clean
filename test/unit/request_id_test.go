package unit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-gin-clean/internal/delivery/http/middleware"

	"github.com/gin-gonic/gin"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("AC-006: Generates UUID When Header Absent", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.RequestID())
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, middleware.GetRequestID(c))
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		headerID := w.Header().Get(middleware.RequestIDHeader)
		if headerID == "" {
			t.Fatalf("expected X-Request-ID response header, got empty")
		}

		bodyID := w.Body.String()
		if headerID != bodyID {
			t.Fatalf("expected context request_id %q to match header %q", bodyID, headerID)
		}
	})

	t.Run("Preserves Existing Incoming X-Request-ID Header", func(t *testing.T) {
		router := gin.New()
		router.Use(middleware.RequestID())
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, middleware.GetRequestID(c))
		})

		incomingID := "custom-trace-id-12345"
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(middleware.RequestIDHeader, incomingID)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Header().Get(middleware.RequestIDHeader) != incomingID {
			t.Fatalf("expected %q, got %q", incomingID, w.Header().Get(middleware.RequestIDHeader))
		}
	})
}
