package permission

type Permission string

const (
	UserRead   Permission = "user:read"
	UserCreate Permission = "user:create"
	UserUpdate Permission = "user:update"
	UserDelete Permission = "user:delete"
)
