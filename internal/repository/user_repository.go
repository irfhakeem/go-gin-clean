package repository

import (
	"context"

	"go-gin-clean/internal/entity"

	"gorm.io/gorm"
)

type UserRepositoryInterface interface {
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

type UserRepository struct {
	db       *gorm.DB
	baseRepo BaseRepository[entity.User]
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	baseRepo := NewBaseRepository[entity.User](db)
	return &UserRepository{
		db:       db,
		baseRepo: *baseRepo,
	}
}

func (r *UserRepository) FindAll(ctx context.Context, limit, offset int, search string) ([]*entity.User, int64, error) {
	return r.baseRepo.FindAll(ctx, limit, offset, "name LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	return r.baseRepo.FindFirst(ctx, "id = ?::uuid", id)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	return r.baseRepo.FindFirst(ctx, "email = ?", email)
}

func (r *UserRepository) ExistByEmail(ctx context.Context, email string) bool {
	isExist, _ := r.baseRepo.WhereExisting(ctx, "email = ?", email)
	return isExist
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	if err := r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?::uuid", user.ID).Updates(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByOAuthID(ctx context.Context, provider, oauthID string) (*entity.User, error) {
	return r.baseRepo.FindFirst(ctx, "oauth_provider = ? AND oauth_id = ?", provider, oauthID)
}

func (r *UserRepository) UpdateOAuthInfo(ctx context.Context, userID string, provider, oauthID string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?::uuid", userID).
		Updates(map[string]interface{}{
			"oauth_provider": provider,
			"oauth_id":       oauthID,
			"is_verified":    true,
		}).Error
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?::uuid", id).Error
}
