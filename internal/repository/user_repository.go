package repository

import (
	"context"
	"fmt"
	"go-gin-clean/internal/entity"
	"time"

	"gorm.io/gorm"
)

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

func (r *UserRepository) FindByCode(ctx context.Context, code string) (*entity.User, error) {
	return r.baseRepo.FindFirst(ctx, "code = ?", code)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	return r.baseRepo.FindFirst(ctx, "email = ?", email)
}

func (r *UserRepository) ExistByEmail(ctx context.Context, email string) bool {
	isExist, _ := r.baseRepo.WhereExisting(ctx, "email = ?", email)
	return isExist
}

func (r *UserRepository) ExistByUsername(ctx context.Context, username string) bool {
	isExist, _ := r.baseRepo.WhereExisting(ctx, "username = ?", username)
	return isExist
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Generate code: U + current date (dd) + current month (mm) + current year (yy) + seq (7 digits), max length 15
	// Example: US17112500001

	// Get today's date and year
	var date, month, year string
	if nowVal := ctx.Value("now"); nowVal != nil {
		if now, ok := nowVal.(func() string); ok && now != nil {
			today := now()
			if len(today) >= 8 {
				date = today[6:8]  // dd
				month = today[4:6] // mm
				year = today[2:4]  // yy
			}
		}
	}
	if date == "" || month == "" || year == "" {
		t := time.Now()
		date = fmt.Sprintf("%02d", t.Day())
		month = fmt.Sprintf("%02d", t.Month())
		year = fmt.Sprintf("%02d", t.Year()%100)
	}

	var lastCode string
	if err := tx.Raw("SELECT last_code FROM sequences WHERE entity_name = 'users' FOR UPDATE").Scan(&lastCode).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	newSeq := 1
	currentDateStr := fmt.Sprintf("%s%s%s", date, month, year)

	if len(lastCode) == 15 && lastCode[:2] == "US" {
		lastDateStr := lastCode[2:8]

		if lastDateStr == currentDateStr {
			var lastSeqInt int
			_, err := fmt.Sscanf(lastCode[8:], "%07d", &lastSeqInt)
			if err == nil {
				newSeq = lastSeqInt + 1
			}
		}
	}

	user.Code = fmt.Sprintf("US%s%07d", currentDateStr, newSeq)

	if err := tx.Exec("UPDATE sequences SET last_code = ?, updated_at = NOW() WHERE entity_name = 'users'", user.Code).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User, code string) (*entity.User, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&entity.User{}).Where("code = ?", code).Updates(user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return user, nil
}

// FindByOAuthID finds a user by OAuth provider and ID
func (r *UserRepository) FindByOAuthID(ctx context.Context, provider, oauthID string) (*entity.User, error) {
	return r.baseRepo.FindFirst(ctx, "oauth_provider = ? AND oauth_id = ?", provider, oauthID)
}

// UpdateOAuthInfo updates OAuth information for an existing user
func (r *UserRepository) UpdateOAuthInfo(ctx context.Context, userID string, provider, oauthID string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?::uuid", userID).
		Updates(map[string]interface{}{
			"oauth_provider": provider,
			"oauth_id":       oauthID,
			"is_verified":    true,
		}).Error
}

func (r *UserRepository) Delete(ctx context.Context, code string) error {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Delete(&entity.User{}, "code = ?", code).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
