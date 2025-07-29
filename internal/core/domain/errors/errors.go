package errors

import "errors"

// Application errors
var (
	ErrUnsupportedFileType     = errors.New("unsupported file type")
	ErrFileTooLarge            = errors.New("file size is too large")
	ErrInvalidInput            = errors.New("invalid input provided")
	ErrAuthHeaderMissing       = errors.New("authorization header is missing")
	ErrTokenInvalid            = errors.New("token is invalid or expired")
	ErrTokenNotFound           = errors.New("token not found")
	ErrTokenExpired            = errors.New("token has expired")
	ErrInvalidClaims           = errors.New("invalid claims in token")
	ErrInvalidIDFormat         = errors.New("invalid ID format")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
)

// Domain errors
var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrPasswordNotMatch      = errors.New("password does not match")
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrInvalidEmailLength    = errors.New("email length must be between 5 and 254 characters")
	ErrInvalidPasswordLength = errors.New("password must be at least 8 characters long")
	ErrPasswordWeak          = errors.New("password must contain at least one special character")
)
