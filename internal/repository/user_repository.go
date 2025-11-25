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

func (r *UserRepository) FindByID(ctx context.Context, pkid int64) (*entity.User, error) {
	return r.baseRepo.FindByID(ctx, pkid)
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

	// Generate code: U + letter (A-Z) + current date (dd) + current year (yy) + seq (5 digits), max length 11
	// Example: UA172500001

	// Find the current letter (A-Z) and sequence
	var letter byte = 'A'
	var seq int = 1

	// Get today's date and year
	var date, year string
	if nowVal := ctx.Value("now"); nowVal != nil {
		if now, ok := nowVal.(func() string); ok && now != nil {
			today := now()
			if len(today) >= 8 {
				date = today[6:8] // dd
				year = today[2:4] // yy
			}
		}
	}
	if date == "" || year == "" {
		t := time.Now()
		date = fmt.Sprintf("%02d", t.Day())
		year = fmt.Sprintf("%02d", t.Year()%100)
	}

	codePrefix := fmt.Sprintf("U%s%s%s", string(letter), date, year)

	var lastCode string
	if err := tx.Raw("SELECT code FROM users WHERE code LIKE ? ORDER BY code DESC LIMIT 1", codePrefix+"%").Scan(&lastCode).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if len(lastCode) == 11 {
		// Extract sequence
		var lastSeq int
		_, err := fmt.Sscanf(lastCode[6:], "%05d", &lastSeq)
		if err == nil {
			seq = lastSeq + 1
			if seq > 99999 {
				// Move to next letter
				letter++
				if letter > 'Z' {
					letter = 'A'
				}
				seq = 1
				codePrefix = fmt.Sprintf("U%s%s%s", string(letter), date, year)
			}
		}
	}

	user.Code = fmt.Sprintf("%s%05d", codePrefix, seq)

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
func (r *UserRepository) UpdateOAuthInfo(ctx context.Context, userID int64, provider, oauthID string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("pkid = ?", userID).
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
