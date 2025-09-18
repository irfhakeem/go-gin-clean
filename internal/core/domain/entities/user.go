package entities

import (
	"go-gin-clean/internal/core/domain/enums"
)

type User struct {
	ID       int64        `json:"id" gorm:"primaryKey;autoIncrement"`
	Name     string       `json:"name" gorm:"not null"`
	Email    string       `json:"email" gorm:"uniqueIndex;not null"`
	Password string       `json:"password" gorm:"not null"`
	Avatar   string       `json:"avatar" gorm:"default:''"`
	Gender   enums.Gender `json:"gender" gorm:"type:gender;default:null"`
	IsActive bool         `json:"is_active" gorm:"default:false;not null"`

	Audit
}

func (User) TableName() string {
	return "users"
}

func NewUser(name, email, password, avatar string, Gender enums.Gender) (*User, error) {
	return &User{
		Name:     name,
		Email:    email,
		Password: password,
		Avatar:   avatar,
		Gender:   Gender,
	}, nil
}

func (u *User) Equals(other *User) bool {
	if other == nil {
		return false
	}

	return u.ID == other.ID &&
		u.Name == other.Name &&
		u.Email == other.Email &&
		u.Password == other.Password &&
		u.Avatar == other.Avatar &&
		u.Gender == other.Gender
}

func (u *User) ChangePassword(newPassword string) error {
	u.Password = newPassword
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
