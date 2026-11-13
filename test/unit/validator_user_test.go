package unit_test

import (
	"strings"
	"testing"

	"go-gin-clean/internal/dto"
	customvalidator "go-gin-clean/internal/dto/validator"

	"github.com/gin-gonic/gin/binding"
)

func TestPasswordValidation(t *testing.T) {
	_ = customvalidator.RegisterCustomValidations()

	t.Run("Valid Password Passes", func(t *testing.T) {
		req := dto.CreateUserRequest{
			Name:     "Valid User",
			Email:    "valid@example.com",
			Password: "StrongPassword123!",
		}
		err := binding.Validator.ValidateStruct(&req)
		if err != nil {
			t.Fatalf("expected valid password to pass, got: %v", err)
		}
	})

	t.Run("Too Short Password Fails", func(t *testing.T) {
		req := dto.CreateUserRequest{
			Name:     "Valid User",
			Email:    "valid@example.com",
			Password: "Short1!",
		}
		err := binding.Validator.ValidateStruct(&req)
		if err == nil {
			t.Fatalf("expected short password (<8 chars) to fail")
		}
	})

	t.Run("Exceeds 72 Characters Fails (Bcrypt Truncation Guard)", func(t *testing.T) {
		longPassword := "StrongPass1!" + strings.Repeat("A", 70)
		req := dto.CreateUserRequest{
			Name:     "Valid User",
			Email:    "valid@example.com",
			Password: longPassword,
		}
		err := binding.Validator.ValidateStruct(&req)
		if err == nil {
			t.Fatalf("expected password exceeding 72 bytes to fail")
		}
	})

	t.Run("Missing Special Character Fails", func(t *testing.T) {
		req := dto.CreateUserRequest{
			Name:     "Valid User",
			Email:    "valid@example.com",
			Password: "NoSpecialChar123",
		}
		err := binding.Validator.ValidateStruct(&req)
		if err == nil {
			t.Fatalf("expected password without symbol to fail")
		}
	})
}
