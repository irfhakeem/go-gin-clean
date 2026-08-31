package postgres

import (
	"context"
	"time"

	"go-gin-clean/internal/domain/entity"

	"gorm.io/gorm"
)

type PostgresRefreshTokenRepository struct {
	db       *gorm.DB
	baseRepo *BaseRepository[entity.RefreshToken]
}

func NewPostgresRefreshTokenRepository(db *gorm.DB) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{
		db:       db,
		baseRepo: NewBaseRepository[entity.RefreshToken](db),
	}
}

func (r *PostgresRefreshTokenRepository) Save(ctx context.Context, token *entity.RefreshToken) error {
	return r.db.WithContext(ctx).Exec(
		"INSERT INTO refresh_tokens (token, user_id, is_revoked, expiry_at, created_at, updated_at) VALUES (?, ?::uuid, ?, ?, ?, ?)",
		token.Token, token.UserID.String(), token.IsRevoked, token.ExpiryAt, token.CreatedAt, token.UpdatedAt,
	).Error
}

func (r *PostgresRefreshTokenRepository) FindByToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	return r.baseRepo.FindFirst(ctx, "token = ? AND is_revoked = ? AND expiry_at > ?", token, false, time.Now())
}

func (r *PostgresRefreshTokenRepository) FindByUserID(ctx context.Context, userID string) ([]*entity.RefreshToken, error) {
	return r.baseRepo.Where(ctx, "user_id = ?::uuid", userID)
}

func (r *PostgresRefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&entity.RefreshToken{}).
		Where("user_id = ?::uuid AND is_revoked = ?", userID, false).
		Update("is_revoked", true).Error
}

func (r *PostgresRefreshTokenRepository) RevokeByToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Model(&entity.RefreshToken{}).
		Where("token = ?", token).
		Update("is_revoked", true).Error
}

func (r *PostgresRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expiry_at < ?", time.Now()).
		Delete(&entity.RefreshToken{}).Error
}

func (r *PostgresRefreshTokenRepository) IsTokenValid(ctx context.Context, token string) bool {
	var count int64
	r.db.WithContext(ctx).Model(&entity.RefreshToken{}).
		Where("token = ? AND is_revoked = ? AND expiry_at > ?", token, false, time.Now()).
		Count(&count)
	return count > 0
}
