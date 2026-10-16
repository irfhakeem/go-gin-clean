package message

type MessageKey string

func (c MessageKey) String() string {
	return string(c)
}

const (
	ErrUnsupportedImageType    MessageKey = "UNSUPPORTED_IMAGE_TYPE"
	ErrAuthHeaderMissing       MessageKey = "AUTH_HEADER_MISSING"
	ErrTokenInvalid            MessageKey = "TOKEN_INVALID"
	ErrTokenNotFound           MessageKey = "TOKEN_NOT_FOUND"
	ErrTokenExpired            MessageKey = "TOKEN_EXPIRED"
	ErrInvalidClaims           MessageKey = "INVALID_CLAIMS"
	ErrUnexpectedSigningMethod MessageKey = "UNEXPECTED_SIGNING_METHOD"

	ErrCacheMiss MessageKey = "CACHE_MISS"

	ErrRefreshToken             MessageKey = "REFRESH_TOKEN"

	ErrPrepareVerificationEmail   MessageKey = "PREPARE_VERIFICATION_EMAIL"
	ErrPrepareForgotPasswordEmail MessageKey = "PREPARE_FORGOT_PASSWORD_EMAIL"

	ErrRegisterFailed        MessageKey = "REGISTER_FAILED"
	ErrLoginFailed           MessageKey = "LOGIN_FAILED"
	ErrUserNotFound          MessageKey = "USER_NOT_FOUND"
	ErrEmailAlreadyExists    MessageKey = "EMAIL_ALREADY_EXISTS"
	ErrEmailNotVerified      MessageKey = "EMAIL_NOT_VERIFIED"
	ErrPasswordNotMatch      MessageKey = "PASSWORD_NOT_MATCH"
	ErrCreateUser            MessageKey = "CREATE_USER"
	ErrUpdateUser            MessageKey = "UPDATE_USER"
	ErrDeleteUser            MessageKey = "DELETE_USER"
	ErrUpdateUserStatus      MessageKey = "UPDATE_USER_STATUS"
	ErrGetAllUsers           MessageKey = "GET_ALL_USERS"
	ErrGetUserInformation    MessageKey = "GET_USER_INFORMATION"
	ErrUpdateUserPassword    MessageKey = "UPDATE_USER_PASSWORD"

	ErrInternalServerError MessageKey = "INTERNAL_SERVER_ERROR"
	ErrInvalidRequestBody  MessageKey = "INVALID_REQUEST_BODY"
	ErrUnauthorized        MessageKey = "UNAUTHORIZED"
	ErrForbidden           MessageKey = "FORBIDDEN"
	ErrValidationFailed    MessageKey = "VALIDATION_FAILED"
	ErrLogoutFailed        MessageKey = "LOGOUT_FAILED"
	ErrVerifyEmailFailed   MessageKey = "VERIFY_EMAIL_FAILED"
	ErrResetPasswordFailed MessageKey = "RESET_PASSWORD_FAILED"

	LoginSuccess                  MessageKey = "LOGIN_SUCCESS"
	RegisterSuccess               MessageKey = "REGISTER_SUCCESS"
	RefreshSuccess                MessageKey = "REFRESH_SUCCESS"
	LogoutSuccess                 MessageKey = "LOGOUT_SUCCESS"
	SendVerificationEmailSuccess  MessageKey = "SEND_VERIFICATION_EMAIL_SUCCESS"
	VerifyEmailSuccess            MessageKey = "VERIFY_EMAIL_SUCCESS"
	SendResetPasswordEmailSuccess MessageKey = "SEND_RESET_PASSWORD_EMAIL_SUCCESS"
	ResetPasswordSuccess          MessageKey = "RESET_PASSWORD_SUCCESS"
	ChangePasswordSuccess         MessageKey = "CHANGE_PASSWORD_SUCCESS"
	CreateUserSuccess             MessageKey = "CREATE_USER_SUCCESS"
	UpdateUserSuccess             MessageKey = "UPDATE_USER_SUCCESS"
	ChangeUserStatusSuccess       MessageKey = "CHANGE_USER_STATUS_SUCCESS"
	GetUserInfoSuccess            MessageKey = "GET_USER_INFO_SUCCESS"
	GetAllUsersSuccess            MessageKey = "GET_ALL_USERS_SUCCESS"
	DeleteUserSuccess             MessageKey = "DELETE_USER_SUCCESS"
)
