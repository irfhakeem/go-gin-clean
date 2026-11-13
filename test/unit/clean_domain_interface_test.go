package unit

import (
	"testing"

	"go-gin-clean/internal/domain/gateway"
	"go-gin-clean/internal/domain/repository"
)

func TestCleanDomainInterfacesCompile(t *testing.T) {
	var _ repository.UserRepository = nil
	var _ repository.RefreshTokenRepository = nil
	var _ repository.OutboxRepository = nil
	var _ repository.TransactionManager = nil
	var _ gateway.Publisher = nil
	var _ gateway.Consumer = nil
	var _ gateway.Mailer = nil
	var _ gateway.Storage = nil
	var _ gateway.Cache = nil
	var _ gateway.TokenMaker = nil
	var _ gateway.Hasher = nil
	var _ gateway.Encryptor = nil
}
