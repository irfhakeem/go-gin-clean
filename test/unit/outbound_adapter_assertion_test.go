package unit

import (
	"testing"

	"go-gin-clean/internal/application/port/outbound"
	"go-gin-clean/internal/infrastructure/cache"
	"go-gin-clean/internal/infrastructure/mailer"
	"go-gin-clean/internal/infrastructure/messaging/rabbitmq"
	"go-gin-clean/internal/infrastructure/repository/postgres"
	"go-gin-clean/internal/infrastructure/security"
	"go-gin-clean/internal/infrastructure/storage"
)

func TestOutboundAdaptersImplementPorts(t *testing.T) {
	if false {
		var _ outbound.UserRepository = postgres.NewPostgresUserRepository(nil)
		var _ outbound.RefreshTokenRepository = postgres.NewPostgresRefreshTokenRepository(nil)
		var _ outbound.OutboxRepository = postgres.NewPostgresOutboxRepository(nil)
		var _ outbound.TransactionManager = postgres.NewGormTransactionManager(nil)
		var _ outbound.Cache = cache.NewRedisCache(nil, nil)
		var _ outbound.Mailer = mailer.NewSMTPMailer(nil)
		var _ outbound.Publisher = rabbitmq.NewPublisher(nil, nil)
		var _ outbound.TokenMaker = security.NewJWTMaker(nil)
		var _ outbound.Hasher = security.NewBcryptHasher()
		var _ outbound.Encryptor = security.NewAESEncryptor(nil)
		var _ outbound.Storage = storage.NewCloudinaryStorage(nil)
	}
}
