package policy

import "github.com/google/uuid"

type ScopeType uint8

const (
	ScopeNone ScopeType = iota
	ScopeAll
	ScopeFiltered
)

type Scope struct {
	Type    ScopeType
	UserIDs []uuid.UUID
}

func NoAccess() Scope {
	return Scope{
		Type: ScopeNone,
	}
}

func FullAccess() Scope {
	return Scope{
		Type: ScopeAll,
	}
}

func ActorScope(ids ...uuid.UUID) Scope {
	return Scope{
		Type:    ScopeFiltered,
		UserIDs: ids,
	}
}
