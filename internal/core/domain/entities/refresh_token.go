package entities

import "time"

type RefreshToken struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int64     `json:"user_id" gorm:"not null"`
	Token     string    `json:"token" gorm:"not null;unique"`
	ExpiryAt  time.Time `json:"expiry_at" gorm:"not null;type:timestamp"`
	IsRevoked bool      `json:"is_revoked" gorm:"default:false;not null"`
	User      User      `json:"user" gorm:"foreignKey:UserID;references:ID"`

	Audit
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

func NewRefreshToken(userID int64, token string, expiryAt time.Time, isRevoked bool, user User) *RefreshToken {
	return &RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiryAt:  expiryAt,
		IsRevoked: isRevoked,
		User:      user,
	}
}

func (rf *RefreshToken) GetID() int64 {
	return rf.ID
}

func (rf *RefreshToken) GetUserID() int64 {
	return rf.UserID
}

func (rf *RefreshToken) GetToken() string {
	return rf.Token
}

func (rf *RefreshToken) IsExpired() bool {
	return time.Now().After(rf.ExpiryAt)
}

func (rf *RefreshToken) Revoke() {
	rf.IsRevoked = true
}

func (rf *RefreshToken) IsValid() bool {
	return !rf.IsRevoked && !rf.IsExpired()
}
