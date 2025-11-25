package repository

import (
	"context"
	"go-gin-clean/internal/entity"
	"time"

	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db       *gorm.DB
	baseRepo BaseRepository[entity.RefreshToken]
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	baseRepo := NewBaseRepository[entity.RefreshToken](db)
	return &RefreshTokenRepository{
		db:       db,
		baseRepo: *baseRepo,
	}
}

func (r *RefreshTokenRepository) Save(ctx context.Context, token *entity.RefreshToken) error {
	return r.db.WithContext(ctx).Exec(
		"INSERT INTO refresh_tokens (token, user_pkid, is_revoked, expiry_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		token.Token, token.UserPKID, token.IsRevoked, token.ExpiryAt, token.CreatedAt, token.UpdatedAt,
	).Error
}

func (r *RefreshTokenRepository) FindByToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	return r.baseRepo.FindFirst(ctx, "token = ? AND is_revoked = ? AND expiry_at > ?", token, false, time.Now())
}

func (r *RefreshTokenRepository) FindByUserID(ctx context.Context, userID int64) ([]*entity.RefreshToken, error) {
	return r.baseRepo.Where(ctx, "user_pkid = ?", userID)
}

func (r *RefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Model(&entity.RefreshToken{}).
		Where("user_pkid = ?", userID).
		Update("is_revoked", true).Error
}

func (r *RefreshTokenRepository) RevokeByToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Model(&entity.RefreshToken{}).
		Where("token = ?", token).
		Update("is_revoked", true).Error
}

func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expiry_at < ?", time.Now()).
		Delete(&entity.RefreshToken{}).Error
}

func (r *RefreshTokenRepository) IsTokenValid(ctx context.Context, token string) bool {
	var count int64
	r.db.WithContext(ctx).Model(&entity.RefreshToken{}).
		Where("token = ? AND is_revoked = ? AND expiry_at > ?", token, false, time.Now()).
		Count(&count)
	return count > 0
}
