package postgres

import (
	"context"

	"go-gin-clean/internal/domain/entity"

	"gorm.io/gorm"
)

type PostgresUserRepository struct {
	db       *gorm.DB
	baseRepo *BaseRepository[entity.User]
}

func NewPostgresUserRepository(db *gorm.DB) *PostgresUserRepository {
	return &PostgresUserRepository{
		db:       db,
		baseRepo: NewBaseRepository[entity.User](db),
	}
}

func (r *PostgresUserRepository) FindAll(ctx context.Context, limit, offset int, search string) ([]*entity.User, int64, error) {
	return r.baseRepo.FindAll(ctx, limit, offset, "name LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	return r.baseRepo.FindFirst(ctx, "id = ?::uuid", id)
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	return r.baseRepo.FindFirst(ctx, "email = ?", email)
}

func (r *PostgresUserRepository) ExistByEmail(ctx context.Context, email string) bool {
	isExist, _ := r.baseRepo.WhereExisting(ctx, "email = ?", email)
	return isExist
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	if err := r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?::uuid", user.ID).Updates(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *PostgresUserRepository) FindByOAuthID(ctx context.Context, provider, oauthID string) (*entity.User, error) {
	return r.baseRepo.FindFirst(ctx, "oauth_provider = ? AND oauth_id = ?", provider, oauthID)
}

func (r *PostgresUserRepository) UpdateOAuthInfo(ctx context.Context, userID string, provider, oauthID string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?::uuid", userID).
		Updates(map[string]interface{}{
			"oauth_provider": provider,
			"oauth_id":       oauthID,
			"is_verified":    true,
		}).Error
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?::uuid", id).Error
}
