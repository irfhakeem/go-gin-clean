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

// Service errors
var (
	ErrGenerateToken   = errors.New("failed to generate token")
	ErrTranslateToken  = errors.New("failed to translate token")
	ErrHashPassword    = errors.New("failed to save password")
	ErrCheckPassword   = errors.New("failed to check password")
	ErrSendEmail       = errors.New("failed to send email")
	ErrUploadFile      = errors.New("failed to upload file")
	ErrDeleteFile      = errors.New("failed to delete file")
	ErrUploadImage     = errors.New("failed to upload image")
	ErrDeleteImage     = errors.New("failed to delete image")
	ErrProcessImage    = errors.New("failed to process image")
	ErrCreateFileSpace = errors.New("failed to create file space")
	ErrToken           = errors.New("token error")
)

// Domain errors
var (
	// User related errors
	ErrUserNotFound           = errors.New("user not found")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrEmailAlreadyExists     = errors.New("email already exists")
	ErrEmailNotVerified       = errors.New("email not verified")
	ErrPasswordNotMatch       = errors.New("password does not match")
	ErrInvalidEmail           = errors.New("invalid email format")
	ErrInvalidEmailLength     = errors.New("email length must be between 5 and 254 characters")
	ErrInvalidPasswordLength  = errors.New("password must be at least 8 characters long")
	ErrPasswordWeak           = errors.New("password must contain at least one special character")
	ErrInvalidGender          = errors.New("gener invalid value")
	ErrInvalidPhone           = errors.New("invalid phone number")
	ErrPhoneAlreadyExists     = errors.New("phone number already exists")
	ErrInvalidUsername        = errors.New("invalid username")
	ErrUsernameAlreadyExists  = errors.New("username already exists")
	ErrSubscriptionNotFound   = errors.New("subscription not found")
	ErrSubscriptionExists     = errors.New("subscription already exists")
	ErrInvalidOAuthProvider   = errors.New("invalid OAuth provider")
	ErrOAuthStateInvalid      = errors.New("invalid OAuth state")
	ErrOAuthCodeExchange      = errors.New("failed to exchange OAuth code for token")
	ErrOAuthUserInfo          = errors.New("failed to get user info from OAuth provider")
	ErrOAuthUserUseOAuthLogin = errors.New("user registered via OAuth, please use OAuth login")
)
