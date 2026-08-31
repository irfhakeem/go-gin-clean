package policy

import (
	"go-gin-clean/internal/domain/vo"

	"github.com/google/uuid"
)

type Actor struct {
	ID   uuid.UUID
	Role vo.Role
}

func NewActor(id, role string) (Actor, error) {
	actorID, err := uuid.Parse(id)
	if err != nil {
		return Actor{}, err
	}

	actorRole, err := vo.ParseRole(role)
	if actorRole == "" {
		return Actor{}, err
	}

	return Actor{
		ID:   actorID,
		Role: actorRole,
	}, nil
}
