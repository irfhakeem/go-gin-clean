package port

import (
	"context"
	"time"

	"go-gin-clean/internal/domain/entity"
)

type UserRepository interface {
	FindAll(ctx context.Context, limit, offset int, search string) ([]*entity.User, int64, error)
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByOAuthID(ctx context.Context, provider, oauthID string) (*entity.User, error)
	ExistByEmail(ctx context.Context, email string) bool
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	UpdateOAuthInfo(ctx context.Context, userID, provider, oauthID string) error
	Delete(ctx context.Context, id string) error
}

type RefreshTokenRepository interface {
	Save(ctx context.Context, token *entity.RefreshToken) error
	FindByToken(ctx context.Context, token string) (*entity.RefreshToken, error)
	FindByUserID(ctx context.Context, userID string) ([]*entity.RefreshToken, error)
	RevokeAllByUserID(ctx context.Context, userID string) error
	RevokeByToken(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
	IsTokenValid(ctx context.Context, token string) bool
}

type OutboxRepository interface {
	Create(ctx context.Context, msg *entity.Outbox) error
	FetchAndLockPending(ctx context.Context, limit int) ([]*entity.Outbox, error)
	MarkPublished(ctx context.Context, pkid int64) error
	RetryOrFail(ctx context.Context, pkid int64, errMsg string) error
	ResetStuck(ctx context.Context, stuckDuration time.Duration) error
}
