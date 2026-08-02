package error

import (
	"fmt"
	"net/http"
)

type ErrCode string

func (c ErrCode) Error() string {
	return string(c)
}

const (
	ErrUnsupportedFileType     ErrCode = "UNSUPPORTED_FILE_TYPE"
	ErrUnsupportedImageType    ErrCode = "UNSUPPORTED_IMAGE_TYPE"
	ErrFileTooLarge            ErrCode = "FILE_TOO_LARGE"
	ErrInvalidInput            ErrCode = "INVALID_INPUT"
	ErrAuthHeaderMissing       ErrCode = "AUTH_HEADER_MISSING"
	ErrTokenInvalid            ErrCode = "TOKEN_INVALID"
	ErrTokenNotFound           ErrCode = "TOKEN_NOT_FOUND"
	ErrTokenExpired            ErrCode = "TOKEN_EXPIRED"
	ErrInvalidClaims           ErrCode = "INVALID_CLAIMS"
	ErrInvalidIDFormat         ErrCode = "INVALID_ID_FORMAT"
	ErrUnexpectedSigningMethod ErrCode = "UNEXPECTED_SIGNING_METHOD"

	ErrCacheMiss ErrCode = "CACHE_MISS"

	ErrGenerateToken            ErrCode = "GENERATE_TOKEN"
	ErrTranslateToken           ErrCode = "TRANSLATE_TOKEN"
	ErrProcessToken             ErrCode = "PROCESS_TOKEN"
	ErrHashPassword             ErrCode = "HASH_PASSWORD"
	ErrCheckPassword            ErrCode = "CHECK_PASSWORD"
	ErrProcessUserPassword      ErrCode = "PROCESS_USER_PASSWORD"
	ErrPasswordNotMeetsCriteria ErrCode = "PASSWORD_NOT_MEETS_CRITERIA"
	ErrSendEmail                ErrCode = "SEND_EMAIL"
	ErrUploadFile               ErrCode = "UPLOAD_FILE"
	ErrDeleteFile               ErrCode = "DELETE_FILE"
	ErrUploadImage              ErrCode = "UPLOAD_IMAGE"
	ErrDeleteImage              ErrCode = "DELETE_IMAGE"
	ErrProcessImage             ErrCode = "PROCESS_IMAGE"
	ErrCreateFileSpace          ErrCode = "CREATE_FILE_SPACE"
	ErrAccessToken              ErrCode = "ACCESS_TOKEN"
	ErrRefreshToken             ErrCode = "REFRESH_TOKEN"

	ErrInvalidOAuthProvider       ErrCode = "INVALID_OAUTH_PROVIDER"
	ErrOAuthStateInvalid          ErrCode = "OAUTH_STATE_INVALID"
	ErrOAuthCodeExchange          ErrCode = "OAUTH_CODE_EXCHANGE"
	ErrOAuthUserInfo              ErrCode = "OAUTH_USER_INFO"
	ErrOAuthUserUseOAuthLogin     ErrCode = "OAUTH_USER_USE_OAUTH_LOGIN"
	ErrOAuthExisting              ErrCode = "OAUTH_EXISTING"
	ErrLinkOAuth                  ErrCode = "LINK_OAUTH"
	ErrOAuthSignUp                ErrCode = "OAUTH_SIGN_UP"
	ErrOAuthCallback              ErrCode = "OAUTH_CALLBACK"
	ErrPrepareVerificationEmail   ErrCode = "PREPARE_VERIFICATION_EMAIL"
	ErrPrepareForgotPasswordEmail ErrCode = "PREPARE_FORGOT_PASSWORD_EMAIL"

	ErrRegisterFailed        ErrCode = "REGISTER_FAILED"
	ErrLoginFailed           ErrCode = "LOGIN_FAILED"
	ErrUserNotFound          ErrCode = "USER_NOT_FOUND"
	ErrTerminateAllSessions  ErrCode = "TERMINATE_ALL_SESSIONS"
	ErrUserAlreadyExists     ErrCode = "USER_ALREADY_EXISTS"
	ErrEmailAlreadyExists    ErrCode = "EMAIL_ALREADY_EXISTS"
	ErrEmailNotVerified      ErrCode = "EMAIL_NOT_VERIFIED"
	ErrPasswordNotMatch      ErrCode = "PASSWORD_NOT_MATCH"
	ErrInvalidEmail          ErrCode = "INVALID_EMAIL"
	ErrInvalidEmailLength    ErrCode = "INVALID_EMAIL_LENGTH"
	ErrInvalidGender         ErrCode = "INVALID_GENDER"
	ErrInvalidPhone          ErrCode = "INVALID_PHONE"
	ErrPhoneAlreadyExists    ErrCode = "PHONE_ALREADY_EXISTS"
	ErrInvalidUsername       ErrCode = "INVALID_USERNAME"
	ErrUsernameAlreadyExists ErrCode = "USERNAME_ALREADY_EXISTS"
	ErrCreateUser            ErrCode = "CREATE_USER"
	ErrUpdateUser            ErrCode = "UPDATE_USER"
	ErrDeleteUser            ErrCode = "DELETE_USER"
	ErrUpdateUserStatus      ErrCode = "UPDATE_USER_STATUS"
	ErrGetAllUsers           ErrCode = "GET_ALL_USERS"
	ErrGetUserInformation    ErrCode = "GET_USER_INFORMATION"
	ErrActivateUser          ErrCode = "ACTIVATE_USER"
	ErrUpdateUserPassword    ErrCode = "UPDATE_USER_PASSWORD"

	ErrInternalServerError ErrCode = "INTERNAL_SERVER_ERROR"
	ErrInvalidRequestBody  ErrCode = "INVALID_REQUEST_BODY"
	ErrUnauthorized        ErrCode = "UNAUTHORIZED"
	ErrValidationFailed    ErrCode = "VALIDATION_FAILED"
	ErrLogoutFailed        ErrCode = "LOGOUT_FAILED"
	ErrVerifyEmailFailed   ErrCode = "VERIFY_EMAIL_FAILED"
	ErrResetPasswordFailed ErrCode = "RESET_PASSWORD_FAILED"
	ErrOAuthLoginFailed    ErrCode = "OAUTH_LOGIN_FAILED"
	ErrOAuthUnlinkFailed   ErrCode = "OAUTH_UNLINK_FAILED"

	LoginSuccess                  ErrCode = "LOGIN_SUCCESS"
	RegisterSuccess               ErrCode = "REGISTER_SUCCESS"
	RefreshSuccess                ErrCode = "REFRESH_SUCCESS"
	LogoutSuccess                 ErrCode = "LOGOUT_SUCCESS"
	SendVerificationEmailSuccess  ErrCode = "SEND_VERIFICATION_EMAIL_SUCCESS"
	VerifyEmailSuccess            ErrCode = "VERIFY_EMAIL_SUCCESS"
	SendResetPasswordEmailSuccess ErrCode = "SEND_RESET_PASSWORD_EMAIL_SUCCESS"
	ResetPasswordSuccess          ErrCode = "RESET_PASSWORD_SUCCESS"
	ChangePasswordSuccess         ErrCode = "CHANGE_PASSWORD_SUCCESS"
	CreateUserSuccess             ErrCode = "CREATE_USER_SUCCESS"
	UpdateUserSuccess             ErrCode = "UPDATE_USER_SUCCESS"
	ChangeUserStatusSuccess       ErrCode = "CHANGE_USER_STATUS_SUCCESS"
	GetUserInfoSuccess            ErrCode = "GET_USER_INFO_SUCCESS"
	GetAllUsersSuccess            ErrCode = "GET_ALL_USERS_SUCCESS"
	DeleteUserSuccess             ErrCode = "DELETE_USER_SUCCESS"
	OAuthLoginSuccess             ErrCode = "OAUTH_LOGIN_SUCCESS"
	OAuthCallbackSuccess          ErrCode = "OAUTH_CALLBACK_SUCCESS"
	OAuthLinkSuccess              ErrCode = "OAUTH_LINK_SUCCESS"
	OAuthUnlinkSuccess            ErrCode = "OAUTH_UNLINK_SUCCESS"
)

type AppError struct {
	HTTPStatus int
	Code       ErrCode
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %v", e.Code, e.Err)
	}
	return string(e.Code)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(httpStatus int, code ErrCode, err error) *AppError {
	return &AppError{
		HTTPStatus: httpStatus,
		Code:       code,
		Err:        err,
	}
}

func BadRequest(code ErrCode, err error) *AppError {
	return NewAppError(http.StatusBadRequest, code, err)
}

func Unauthorized(code ErrCode, err error) *AppError {
	return NewAppError(http.StatusUnauthorized, code, err)
}

func Forbidden(code ErrCode, err error) *AppError {
	return NewAppError(http.StatusForbidden, code, err)
}

func NotFound(code ErrCode, err error) *AppError {
	return NewAppError(http.StatusNotFound, code, err)
}

func InternalServerError(code ErrCode, err error) *AppError {
	return NewAppError(http.StatusInternalServerError, code, err)
}
