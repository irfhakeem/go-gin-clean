package policy

import (
	"go-gin-clean/internal/domain/vo"
)

type UserPolicy interface {
	Scope(actor Actor, action Action) Scope
}

type userPolicy struct{}

func NewUserPolicy() UserPolicy {
	return &userPolicy{}
}

func (p *userPolicy) Scope(actor Actor, action Action) Scope {
	switch actor.Role {
	case vo.RoleSuperAdmin:
		return FullAccess()
	case vo.RoleUser:
		switch action {
		case ActionRead:
			return FilteredScope(actor.ID)
		case ActionUpdate:
			return FilteredScope(actor.ID)
		case ActionDelete:
			return FilteredScope(actor.ID)
		default:
			return NoAccess()
		}
	default:
		return NoAccess()
	}
}
