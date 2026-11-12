package unit_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-gin-clean/internal/delivery/http/route"

	"github.com/gin-gonic/gin"
)

func TestHealthAndReadyProbes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	route.SetupRoutes(router, nil, nil, nil, nil)

	t.Run("AC-008: Liveness Probe /health returns 200 OK with Request ID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", w.Code)
		}

		var body map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		if body["status"] != "healthy" {
			t.Fatalf("expected status healthy, got %v", body["status"])
		}
		if body["request_id"] == nil || body["request_id"] == "" {
			t.Fatalf("expected request_id in health response, got %v", body["request_id"])
		}
	})

	t.Run("AC-008: Readiness Probe /ready returns 200 OK", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/ready", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", w.Code)
		}

		var body map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		if body["status"] != "ready" {
			t.Fatalf("expected status ready, got %v", body["status"])
		}
	})
}
