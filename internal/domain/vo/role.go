package vo

import (
	"errors"
)

var ErrInvalidRole = errors.New("invalid role")

type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleUser       Role = "user"
)

func (r Role) String() string {
	return string(r)
}

func ParseRole(role string) (Role, error) {
	switch Role(role) {
	case RoleSuperAdmin:
		return RoleSuperAdmin, nil
	case RoleUser:
		return RoleUser, nil
	default:
		return "", ErrInvalidRole
	}
}

func IsValidRole(role string) bool {
	switch Role(role) {
	case RoleSuperAdmin, RoleUser:
		return true
	default:
		return false
	}
}
