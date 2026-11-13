package mock_test

import (
	"testing"

	"go-gin-clean/internal/domain/gateway"
	"go-gin-clean/internal/domain/repository"
	"go-gin-clean/test/mock"
)

func TestMocksImplementDomainInterfaces(t *testing.T) {
	var _ repository.UserRepository = (*mock.MockUserRepository)(nil)
	var _ repository.RefreshTokenRepository = (*mock.MockRefreshTokenRepository)(nil)
	var _ repository.OutboxRepository = (*mock.MockOutboxRepository)(nil)
	var _ repository.TransactionManager = (*mock.MockTransactionManager)(nil)
	var _ gateway.Cache = (*mock.MockCache)(nil)
	var _ gateway.Mailer = (*mock.MockMailer)(nil)
	var _ gateway.Publisher = (*mock.MockPublisher)(nil)
	var _ gateway.TokenMaker = (*mock.MockTokenMaker)(nil)
	var _ gateway.Hasher = (*mock.MockHasher)(nil)
	var _ gateway.Encryptor = (*mock.MockEncryptor)(nil)
	var _ gateway.Storage = (*mock.MockStorage)(nil)
}
