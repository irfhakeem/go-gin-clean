package entity

import (
	"go-gin-clean/internal/domain/vo"
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;column:user_id"`
	Token     string    `gorm:"not null;unique"`
	ExpiryAt  time.Time `gorm:"not null;type:timestamp"`
	IsRevoked bool      `gorm:"default:false;not null"`

	User *User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	vo.Audit
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

func NewRefreshToken(userID uuid.UUID, token string, expiryAt time.Time, isRevoked bool, user *User) *RefreshToken {
	return &RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiryAt:  expiryAt,
		IsRevoked: isRevoked,
		User:      user,
	}
}
