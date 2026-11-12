package unit

import (
	"testing"

	"go-gin-clean/internal/application/port/inbound"
)

func TestInboundPortsCompile(t *testing.T) {
	var _ inbound.AuthUseCase = nil
	var _ inbound.UserUseCase = nil
	var _ inbound.EmailUseCase = nil
	var _ inbound.OutboxUseCase = nil
}
