package unit_test

import (
	"context"
	"errors"
	"testing"

	"go-gin-clean/internal/usecase"
	"go-gin-clean/internal/domain/entity"
	"go-gin-clean/internal/dto"
	"go-gin-clean/pkg/config"
	"go-gin-clean/test/mock"

	"github.com/google/uuid"
)

func TestAuthUseCase_OutboxAndSessionHardening(t *testing.T) {
	cfg := &config.ServerConfig{
		AppUrl: "http://localhost:8080",
	}

	t.Run("AC-001: Register Outbox Atomicity Fails Cleanly If Outbox Save Fails", func(t *testing.T) {
		txMgr := mock.NewMockTransactionManager()
		userRepo := mock.NewMockUserRepository()
		refreshRepo := mock.NewMockRefreshTokenRepository()
		outboxRepo := mock.NewMockOutboxRepository()
		outboxRepo.CreateErr = errors.New("outbox write failure")

		pub := mock.NewMockPublisher()
		outboxUC := usecase.NewOutboxUseCase(outboxRepo, pub, &config.RabbitMQConfig{})

		authUC := usecase.NewAuthUseCase(
			txMgr,
			userRepo,
			refreshRepo,
			outboxUC,
			mock.NewMockTokenMaker(),
			mock.NewMockHasher(),
			mock.NewMockEncryptor(),
			cfg,
		)

		err := authUC.Register(context.Background(), &dto.RegisterRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "Password123!",
		})

		if err == nil {
			t.Fatalf("expected Register to fail when outbox save fails, got nil")
		}

		if !txMgr.Called {
			t.Fatalf("expected txManager.WithTx to be called")
		}

		if !txMgr.RolledBack {
			t.Fatalf("AC-001 violation: expected transaction rollback when outbox write fails")
		}
	})

	t.Run("AC-005: ResetPassword Revokes Active Refresh Tokens", func(t *testing.T) {
		txMgr := mock.NewMockTransactionManager()
		targetUser, _ := entity.NewUser("Test", "test@example.com", "oldhash")
		targetUser.ID = uuid.New()

		userRepo := mock.NewMockUserRepository()
		userRepo.Users[targetUser.ID.String()] = targetUser

		refreshRepo := mock.NewMockRefreshTokenRepository()
		outboxRepo := mock.NewMockOutboxRepository()
		pub := mock.NewMockPublisher()
		outboxUC := usecase.NewOutboxUseCase(outboxRepo, pub, &config.RabbitMQConfig{})

		encryptor := mock.NewMockEncryptor()
		// Setup mock token decryption returning user's email and future expiry
		token, _ := encryptor.EncryptURLSafe("test@example.com_2099-01-01T00:00:00Z")

		authUC := usecase.NewAuthUseCase(
			txMgr,
			userRepo,
			refreshRepo,
			outboxUC,
			mock.NewMockTokenMaker(),
			mock.NewMockHasher(),
			encryptor,
			cfg,
		)

		err := authUC.ResetPassword(context.Background(), &dto.ResetPasswordRequest{
			Token:       token,
			NewPassword: "NewStrongPassword123!",
		})

		if err != nil {
			t.Fatalf("unexpected error on reset password: %v", err)
		}

		if !txMgr.Called {
			t.Fatalf("expected txManager.WithTx to be called on ResetPassword")
		}

		foundRevoked := false
		for _, id := range refreshRepo.RevokedUsers {
			if id == targetUser.ID.String() {
				foundRevoked = true
				break
			}
		}

		if !foundRevoked {
			t.Fatalf("AC-005 violation: expected RevokeAllByUserID(%q), revoked list was %v", targetUser.ID.String(), refreshRepo.RevokedUsers)
		}
	})
}
