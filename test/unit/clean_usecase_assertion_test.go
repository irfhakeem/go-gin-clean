package unit

import (
	"testing"

	"go-gin-clean/internal/usecase"
)

func TestCleanUseCaseBoundaries(t *testing.T) {
	var _ usecase.AuthUseCase = nil
	var _ usecase.UserUseCase = nil
	var _ usecase.EmailUseCase = nil
	var _ usecase.OutboxUseCase = nil
}
