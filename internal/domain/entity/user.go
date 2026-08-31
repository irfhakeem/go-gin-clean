package entity

import (
	"go-gin-clean/internal/domain/vo"

	"github.com/google/uuid"
)

const (
	EventUserRegistered    = "user.register"
	EventUserResetPassword = "user.reset_password"
)

type (
	User struct {
		ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id"`
		Name       string    `gorm:"not null"`
		Email      string    `gorm:"type:varchar(100);uniqueIndex;not null;column:email"`
		Password   string    `gorm:"type:varchar(255);column:password"`
		Avatar     string    `gorm:"default:''"`
		Gender     vo.Gender `gorm:"type:gender;default:'not_to_say';not null"`
		Role       vo.Role   `gorm:"type:role;default:'User'"`
		IsActive   bool      `gorm:"default:true;not null"`
		IsVerified bool      `gorm:"default:false;not null"`

		OAuthProvider string `gorm:"type:varchar(50);column:oauth_provider"`
		OAuthID       string `gorm:"type:varchar(255);column:oauth_id"`

		Audit
	}

	UserEvent struct {
		UserID uuid.UUID `json:"user_id"`
		Name   string    `json:"name"`
	}

	UserRegisterEvent struct {
		UserEvent
		Email           string `json:"email"`
		VerificationURL string `json:"verification_url"`
	}

	UserResetPasswordEvent struct {
		UserEvent
		Email    string `json:"email"`
		ResetURL string `json:"reset_url"`
	}
)

func (User) TableName() string {
	return "users"
}

func NewUser(name, email, password string) (*User, error) {
	return &User{
		Name:     name,
		Email:    email,
		Password: password,
	}, nil
}

func NewUserFromOAuth(name, emailStr, oauthProvider, oauthID, avatar string) (*User, error) {
	return &User{
		Name:          name,
		Email:         emailStr,
		Password:      "",
		Avatar:        avatar,
		OAuthProvider: oauthProvider,
		OAuthID:       oauthID,
		IsVerified:    true,
		Gender:        vo.GenderNotToSay,
		IsActive:      true,
	}, nil
}

func (u *User) SetPassword(hashedPassword string) {
	u.Password = hashedPassword
}

func (u *User) Activate() {
	u.IsActive = true
}

func (u *User) Deactivate() {
	u.IsActive = false
}

func (u *User) VerifyEmail() {
	u.IsVerified = true
}

func (u *User) IsOAuthUser() bool {
	return u.OAuthProvider != "" && u.OAuthID != ""
}

func (u *User) IsActiveUser() bool {
	return u.IsActive
}

func (u *User) IsVerifiedUser() bool {
	return u.IsVerified
}
