package database

import (
	"context"
	"go-gin-clean/internal/core/domain/entities"
	"go-gin-clean/internal/core/ports"

	"gorm.io/gorm"
)

type UserRepository struct {
	db       *gorm.DB
	baseRepo ports.BaseRepository[entities.User]
}

func NewUserRepository(db *gorm.DB) ports.UserRepository {
	baseRepo := NewBaseRepository[entities.User](db)
	return &UserRepository{
		db:       db,
		baseRepo: baseRepo,
	}
}

func (r *UserRepository) FindAll(ctx context.Context, limit, offset int, search string) ([]*entities.User, int64, error) {
	return r.baseRepo.FindAll(ctx, limit, offset, "name LIKE ? OR email_email LIKE ?", "%"+search+"%", "%"+search+"%")
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*entities.User, error) {
	return r.baseRepo.FindByID(ctx, id)
}

func (r *UserRepository) Create(ctx context.Context, user *entities.User) (*entities.User, error) {
	return r.baseRepo.Create(ctx, user)
}

func (r *UserRepository) Update(ctx context.Context, user *entities.User) (*entities.User, error) {
	return r.baseRepo.Update(ctx, user)
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return r.baseRepo.Delete(ctx, id)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	return r.baseRepo.FindFirst(ctx, "email_email = ?", email)
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) bool {
	isExist, _ := r.baseRepo.WhereExisting(ctx, "email_email = ?", email)
	return isExist
}
