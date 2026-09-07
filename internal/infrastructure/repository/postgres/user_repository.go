package postgres

import (
	"context"

	"go-gin-clean/internal/application/port"
	"go-gin-clean/internal/domain/entity"

	"gorm.io/gorm"
)

type PostgresUserRepository struct {
	db       *gorm.DB
	baseRepo *BaseRepository[entity.User]
}

func NewPostgresUserRepository(db *gorm.DB) port.UserRepository {
	return &PostgresUserRepository{
		db:       db,
		baseRepo: NewBaseRepository[entity.User](db),
	}
}

func (r *PostgresUserRepository) FindAll(ctx context.Context, params port.FindAllUsersParams) ([]*entity.User, int64, error) {
	var users []*entity.User
	var count int64

	var user entity.User
	q := r.db.WithContext(ctx).Model(&user)

	if params.IDs != nil {
		if len(params.IDs) == 0 {
			q = q.Where("1 = 0")
		} else {
			q = q.Where("id IN ?", params.IDs)
		}
	}

	if params.Search != "" {
		searchPattern := "%" + params.Search + "%"
		q = q.Where("name ILIKE ? OR email ILIKE ?", searchPattern, searchPattern)
	}

	if params.Role != "" {
		q = q.Where("role = ?", params.Role)
	}

	if params.IsActive != nil {
		q = q.Where("is_active = ?", *params.IsActive)
	}

	if params.SortBy != "" && (params.Sort == "asc" || params.Sort == "desc") {
		q = q.Order(params.SortBy + " " + params.Sort)
	} else {
		q = q.Order("created_at desc")
	}

	if err := q.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if err := q.Limit(params.Limit).Offset(params.Offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, count, nil
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
	return r.baseRepo.Create(ctx, user)
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	return r.baseRepo.Update(ctx, user, user.ID.String())
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
	return r.baseRepo.Delete(ctx, id)
}
