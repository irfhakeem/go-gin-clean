package entities

import (
	"go-gin-clean/internal/core/domain/enums"
	vo "go-gin-clean/internal/core/domain/valueobjects"
	"time"
)

type User struct {
	ID       int64            `json:"id" gorm:"primaryKey;autoIncrement"`
	Name     string           `json:"name" gorm:"not null"`
	Email    *vo.EmailAddress `json:"email" gorm:"embedded;embeddedPrefix:email_"`
	Password *vo.Password     `json:"password" gorm:"embedded;embeddedPrefix:password_"`
	Avatar   string           `json:"avatar" gorm:"default:''"`
	Gender   enums.Gender     `json:"gender" gorm:"type:gender;default:'Unknown';not null"`
	IsActive bool             `json:"is_active" gorm:"default:true;not null"`

	Audit
}

func NewUser(id int64, name, emailStr, passwordStr, avatar string, Gender enums.Gender, createdAt, updatedAt time.Time) (*User, error) {
	email, err := vo.NewEmailAddress(emailStr)
	if err != nil {
		return nil, err
	}

	password, err := vo.NewPassword(passwordStr)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       id,
		Name:     name,
		Email:    email,
		Password: password,
		Avatar:   avatar,
		Gender:   Gender,
		Audit: Audit{
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			DeletedAt: nil,
			IsDeleted: false,
		},
	}, nil
}

func (u *User) Equals(other *User) bool {
	if other == nil {
		return false
	}

	return u.ID == other.ID &&
		u.Name == other.Name &&
		u.Email.Equals(other.Email) &&
		u.Password.Equals(other.Password) &&
		u.Avatar == other.Avatar &&
		u.Gender == other.Gender
}

func (u *User) ChangePassword(newPassword string) error {
	password, err := vo.NewPassword(newPassword)
	if err != nil {
		return err
	}
	u.Password = password
	return nil
}

func (u *User) UpdateProfile(name, avatar string, gender enums.Gender) {
	if name != "" {
		u.Name = name
	}
	if avatar != "" {
		u.Avatar = avatar
	}
	u.Gender = gender
}

func (u *User) Activate() {
	u.IsActive = true
}

func (u *User) Deactivate() {
	u.IsActive = false
}
