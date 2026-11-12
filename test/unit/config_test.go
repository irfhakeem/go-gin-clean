package unit_test

import (
	"os"
	"testing"

	"go-gin-clean/pkg/config"
)

func TestConfig_ProductionValidation(t *testing.T) {
	t.Run("AC-011: Fails In Production When Default Placeholders Used", func(t *testing.T) {
		_ = os.Setenv("ENVIRONMENT", "production")
		_ = os.Setenv("JWT_ACCESS_SECRET", "your-access-secret-key")
		defer func() {
			_ = os.Unsetenv("ENVIRONMENT")
			_ = os.Unsetenv("JWT_ACCESS_SECRET")
		}()

		_, err := config.Load()
		if err == nil {
			t.Fatalf("expected config.Load() to fail in production with placeholder secret, but got nil")
		}
	})

	t.Run("Succeeds in Development With Default Placeholders", func(t *testing.T) {
		_ = os.Setenv("ENVIRONMENT", "development")
		defer func() {
			_ = os.Unsetenv("ENVIRONMENT")
		}()

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected error in development mode: %v", err)
		}
		if cfg == nil {
			t.Fatalf("expected non-nil config")
		}
	})
}
