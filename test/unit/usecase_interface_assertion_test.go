package unit

import (
	"testing"

	"go-gin-clean/internal/application/port/inbound"
	"go-gin-clean/internal/application/usecase"
)

func TestUseCaseImplementsInboundPorts(t *testing.T) {
	var _ inbound.AuthUseCase = (usecase.AuthUseCase)(nil)
	var _ inbound.UserUseCase = (usecase.UserUseCase)(nil)
	var _ inbound.EmailUseCase = (usecase.EmailUseCase)(nil)
	var _ inbound.OutboxUseCase = (usecase.OutboxUseCase)(nil)
}
