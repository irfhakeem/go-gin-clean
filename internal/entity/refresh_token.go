package entity

import "time"

type RefreshToken struct {
	PKID      int64     `gorm:"primaryKey;autoIncrement;column:pkid"`
	UserPKID  int64     `gorm:"not null;column:user_pkid"`
	Token     string    `gorm:"not null;unique"`
	ExpiryAt  time.Time `gorm:"not null;type:timestamp"`
	IsRevoked bool      `gorm:"default:false;not null"`

	User User `gorm:"foreignKey:UserPKID;references:PKID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Audit
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

func NewRefreshToken(userID int64, token string, expiryAt time.Time, isRevoked bool, user User) *RefreshToken {
	return &RefreshToken{
		UserPKID:  userID,
		Token:     token,
		ExpiryAt:  expiryAt,
		IsRevoked: isRevoked,
		User:      user,
	}
}

func (rf *RefreshToken) GetID() int64 {
	return rf.PKID
}

func (rf *RefreshToken) GetUserID() int64 {
	return rf.UserPKID
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
