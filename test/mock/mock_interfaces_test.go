package mock_test

import (
	"testing"

	"go-gin-clean/internal/application/port/outbound"
	"go-gin-clean/test/mock"
)

func TestMocksImplementOutboundPorts(t *testing.T) {
	var _ outbound.UserRepository = (*mock.MockUserRepository)(nil)
	var _ outbound.RefreshTokenRepository = (*mock.MockRefreshTokenRepository)(nil)
	var _ outbound.OutboxRepository = (*mock.MockOutboxRepository)(nil)
	var _ outbound.TransactionManager = (*mock.MockTransactionManager)(nil)
	var _ outbound.Cache = (*mock.MockCache)(nil)
	var _ outbound.Mailer = (*mock.MockMailer)(nil)
	var _ outbound.Publisher = (*mock.MockPublisher)(nil)
	var _ outbound.TokenMaker = (*mock.MockTokenMaker)(nil)
	var _ outbound.Hasher = (*mock.MockHasher)(nil)
	var _ outbound.Encryptor = (*mock.MockEncryptor)(nil)
	var _ outbound.Storage = (*mock.MockStorage)(nil)
}
