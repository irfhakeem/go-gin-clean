package postgres

import (
	"context"
	"time"

	"go-gin-clean/internal/application/port"
	"go-gin-clean/internal/domain/entity"

	"gorm.io/gorm"
)

type PostgresRefreshTokenRepository struct {
	db *gorm.DB
}

func NewPostgresRefreshTokenRepository(db *gorm.DB) port.RefreshTokenRepository {
	return &PostgresRefreshTokenRepository{db: db}
}

func (r *PostgresRefreshTokenRepository) Save(ctx context.Context, token *entity.RefreshToken) error {
	return getTx(ctx, r.db).Exec(
		"INSERT INTO refresh_tokens (token, user_id, is_revoked, expiry_at, created_at, updated_at) VALUES (?, ?::uuid, ?, ?, ?, ?)",
		token.Token, token.UserID.String(), token.IsRevoked, token.ExpiryAt, token.CreatedAt, token.UpdatedAt,
	).Error
}
func (r *PostgresRefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	return getTx(ctx, r.db).Model(&entity.RefreshToken{}).
		Where("user_id = ?::uuid AND is_revoked = ?", userID, false).
		Update("is_revoked", true).Error
}

func (r *PostgresRefreshTokenRepository) RevokeByToken(ctx context.Context, token string) error {
	return getTx(ctx, r.db).Model(&entity.RefreshToken{}).
		Where("token = ?", token).
		Update("is_revoked", true).Error
}

func (r *PostgresRefreshTokenRepository) IsTokenValid(ctx context.Context, token string) bool {
	var count int64
	getTx(ctx, r.db).Model(&entity.RefreshToken{}).
		Where("token = ? AND is_revoked = ? AND expiry_at > ?", token, false, time.Now()).
		Count(&count)
	return count > 0
}
