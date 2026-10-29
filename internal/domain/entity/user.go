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

		vo.Audit
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

func (u *User) SetPassword(hashedPassword string) {
	u.Password = hashedPassword
}

func (u *User) VerifyEmail() {
	u.IsVerified = true
}

