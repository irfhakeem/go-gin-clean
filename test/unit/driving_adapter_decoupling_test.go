package unit

import (
	"testing"

	"go-gin-clean/internal/application/port/inbound"
	"go-gin-clean/internal/delivery/http"
	"go-gin-clean/internal/delivery/mq"
	"go-gin-clean/internal/delivery/worker"
)

func TestDrivingAdaptersAcceptInboundPorts(t *testing.T) {
	var authPort inbound.AuthUseCase
	var userPort inbound.UserUseCase
	var emailPort inbound.EmailUseCase
	var outboxPort inbound.OutboxUseCase

	_ = http.NewAuthHandler(authPort)
	_ = http.NewUserHandler(userPort)
	_ = mq.NewEmailEventHandler(emailPort)
	_ = worker.NewOutboxWorker(outboxPort)
}
