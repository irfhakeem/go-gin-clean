package permission

import (
	"go-gin-clean/internal/domain/policy"
	"go-gin-clean/internal/domain/vo"
)

type Checker interface {
	HasPermission(actor policy.Actor, permission Permission) bool
}

type checker struct {
	permissions map[vo.Role]map[Permission]bool
}

func NewChecker() Checker {
	permMap := make(map[vo.Role]map[Permission]bool)
	permMap[vo.RoleSuperAdmin] = map[Permission]bool{
		UserRead:   true,
		UserCreate: true,
		UserUpdate: true,
		UserDelete: true,
	}
	permMap[vo.RoleUser] = map[Permission]bool{
		UserRead:   true,
		UserCreate: false,
		UserUpdate: true,
		UserDelete: false,
	}
	return &checker{
		permissions: permMap,
	}
}

func (pc *checker) HasPermission(actor policy.Actor, permission Permission) bool {
	return pc.permissions[actor.Role][permission]
}
