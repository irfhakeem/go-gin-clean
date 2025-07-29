package ports

import (
	"context"
	"go-gin-clean/internal/core/domain/entities"
)

// Repository interfaces (secondary ports)
type BaseRepository[T any] interface {
	FindAll(ctx context.Context, limit, offset int, query any, args ...any) ([]*T, int64, error)
	FindByID(ctx context.Context, id int64) (*T, error)
	FindFirst(ctx context.Context, query any, args ...any) (*T, error)
	Where(ctx context.Context, query any, args ...any) ([]*T, error)
	WhereExisting(ctx context.Context, query any, args ...any) (bool, error)
	Create(ctx context.Context, entity *T) (*T, error)
	Update(ctx context.Context, entity *T) (*T, error)
	Delete(ctx context.Context, id int64) error
}

type UserRepository interface {
	FindAll(ctx context.Context, limit, offset int, search string) ([]*entities.User, int64, error)
	FindByID(ctx context.Context, id int64) (*entities.User, error)
	Create(ctx context.Context, user *entities.User) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) (*entities.User, error)
	Delete(ctx context.Context, id int64) error
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	ExistsByEmail(ctx context.Context, email string) bool
}

type RefreshTokenRepository interface {
	Save(ctx context.Context, token *entities.RefreshToken) error
	FindByToken(ctx context.Context, token string) (*entities.RefreshToken, error)
	FindByUserID(ctx context.Context, userID int64) ([]*entities.RefreshToken, error)
	RevokeAllByUserID(ctx context.Context, userID int64) error
	RevokeByToken(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
	IsTokenValid(ctx context.Context, token string) bool
}
