package policy

import (
	"go-gin-clean/internal/domain/vo"

	"github.com/google/uuid"
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
		return ActorScope(actor.ID)
	default:
		return NoAccess()
	}
}

func userScope(actor Actor, action Action) Scope {
	switch action {
	case ActionRead:
		return ActorScope(
			uuid.Nil,
			actor.ID,
		)

	case ActionUpdate:
		return ActorScope(
			actor.ID,
		)

	default:
		return NoAccess()
	}
}
