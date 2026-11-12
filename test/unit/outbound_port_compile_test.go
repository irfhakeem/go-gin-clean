package unit

import (
	"testing"

	"go-gin-clean/internal/application/port/outbound"
)

func TestOutboundPortsCompile(t *testing.T) {
	var _ outbound.UserRepository = nil
	var _ outbound.RefreshTokenRepository = nil
	var _ outbound.OutboxRepository = nil
	var _ outbound.TransactionManager = nil
	var _ outbound.Publisher = nil
	var _ outbound.Consumer = nil
	var _ outbound.Mailer = nil
	var _ outbound.Storage = nil
	var _ outbound.Cache = nil
	var _ outbound.TokenMaker = nil
	var _ outbound.Hasher = nil
	var _ outbound.Encryptor = nil
}
