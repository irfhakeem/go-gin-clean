package database

import (
	"context"
	"go-gin-clean/internal/core/domain/entities"
	"go-gin-clean/internal/core/ports"
	"time"

	"gorm.io/gorm"
)

type RefreshTokenRepository struct {
	db       *gorm.DB
	baseRepo ports.BaseRepository[entities.RefreshToken]
}

func NewRefreshTokenRepository(db *gorm.DB) ports.RefreshTokenRepository {
	baseRepo := NewBaseRepository[entities.RefreshToken](db)
	return &RefreshTokenRepository{
		db:       db,
		baseRepo: baseRepo,
	}
}

func (r *RefreshTokenRepository) Save(ctx context.Context, token *entities.RefreshToken) error {
	_, err := r.baseRepo.Create(ctx, token)
	return err
}

func (r *RefreshTokenRepository) FindByToken(ctx context.Context, token string) (*entities.RefreshToken, error) {
	return r.baseRepo.FindFirst(ctx, "token = ? AND is_revoked = ? AND expiry_at > ?", token, false, time.Now())
}

func (r *RefreshTokenRepository) FindByUserID(ctx context.Context, userID int64) ([]*entities.RefreshToken, error) {
	return r.baseRepo.Where(ctx, "user_id = ?", userID)
}

func (r *RefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Model(&entities.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("is_revoked", true).Error
}

func (r *RefreshTokenRepository) RevokeByToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Model(&entities.RefreshToken{}).
		Where("token = ?", token).
		Update("is_revoked", true).Error
}

func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expiry_at < ?", time.Now()).
		Delete(&entities.RefreshToken{}).Error
}

func (r *RefreshTokenRepository) IsTokenValid(ctx context.Context, token string) bool {
	var count int64
	r.db.WithContext(ctx).Model(&entities.RefreshToken{}).
		Where("token = ? AND is_revoked = ? AND expiry_at > ?", token, false, time.Now()).
		Count(&count)
	return count > 0
}
