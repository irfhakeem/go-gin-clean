package errors

import "errors"

// Application errors
var (
	ErrUnsupportedFileType     = errors.New("the file type you uploaded is not supported")
	ErrUnsupportedImageType    = errors.New("only JPG, JPEG, and PNG image formats are allowed")
	ErrFileTooLarge            = errors.New("the file size exceeds the maximum allowed limit")
	ErrInvalidInput            = errors.New("the information provided is invalid or incomplete")
	ErrAuthHeaderMissing       = errors.New("authentication is required to access this resource")
	ErrTokenInvalid            = errors.New("your session is invalid or has expired, please log in again")
	ErrTokenNotFound           = errors.New("authentication token not found, please log in")
	ErrTokenExpired            = errors.New("your session has expired, please log in again")
	ErrInvalidClaims           = errors.New("authentication credentials are invalid")
	ErrInvalidIDFormat         = errors.New("the provided identifier format is invalid")
	ErrUnexpectedSigningMethod = errors.New("authentication method is not recognized")
)

// Service errors
var (
	ErrGenerateToken            = errors.New("unable to generate authentication token, please try again")
	ErrTranslateToken           = errors.New("unable to process authentication token")
	ErrProcessToken             = errors.New("unable to process your authentication token, please try again")
	ErrHashPassword             = errors.New("unable to securely process your password")
	ErrCheckPassword            = errors.New("unable to verify your password")
	ErrProcessUserPassword      = errors.New("unable to process your password securely")
	ErrPasswordNotMeetsCriteria = errors.New("password must be at least 8 characters long and contain uppercase letters, lowercase letters, and numbers")
	ErrSendEmail                = errors.New("unable to send email at this time, please try again later")
	ErrUploadFile               = errors.New("unable to upload your file, please try again")
	ErrDeleteFile               = errors.New("unable to delete the file, please try again")
	ErrUploadImage              = errors.New("unable to upload your image, please try again")
	ErrDeleteImage              = errors.New("unable to delete the image, please try again")
	ErrProcessImage             = errors.New("unable to process your image, please ensure it's a valid image file")
	ErrCreateFileSpace          = errors.New("unable to create storage space for your files")
	ErrAccessToken              = errors.New("unable to generate access token, please try logging in again")
	ErrRefreshToken             = errors.New("unable to refresh your session, please log in again")

	// OAuth errors
	ErrInvalidOAuthProvider   = errors.New("the login provider you selected is not supported")
	ErrOAuthStateInvalid      = errors.New("login request is invalid or has expired, please try again")
	ErrOAuthCodeExchange      = errors.New("unable to complete login with the provider, please try again")
	ErrOAuthUserInfo          = errors.New("unable to retrieve your information from the login provider")
	ErrOAuthUserUseOAuthLogin = errors.New("this account is linked to social login, please use the appropriate login method")
	ErrOAuthExisting          = errors.New("this social account is already linked to another user")
	ErrLinkOAuth              = errors.New("unable to link your social account at this time")
	ErrOAuthSignUp            = errors.New("unable to complete registration with social login, please try again")
	ErrOAuthCallback          = errors.New("unable to complete social login, please try again")

	ErrPrepareVerificationEmail   = errors.New("unable to prepare verification email, please try again")
	ErrPrepareForgotPasswordEmail = errors.New("unable to prepare password reset email, please try again")
)

// Domain errors
var (
	// User related errors
	ErrRegisterFailed        = errors.New("unable to complete registration, please try again")
	ErrLoginFailed           = errors.New("unable to log in, please check your credentials")
	ErrUserNotFound          = errors.New("user account not found or is inactive")
	ErrTerminateAllSessions  = errors.New("unable to terminate all active sessions, please try again")
	ErrUserAlreadyExists     = errors.New("an account with this information already exists")
	ErrEmailAlreadyExists    = errors.New("an account with this email address already exists")
	ErrEmailNotVerified      = errors.New("your email address has not been verified, please check your inbox")
	ErrPasswordNotMatch      = errors.New("the password you entered is incorrect")
	ErrInvalidEmail          = errors.New("the email address format is invalid")
	ErrInvalidEmailLength    = errors.New("email address must be between 5 and 254 characters")
	ErrInvalidGender         = errors.New("the gender value provided is invalid")
	ErrInvalidPhone          = errors.New("the phone number format is invalid")
	ErrPhoneAlreadyExists    = errors.New("this phone number is already registered")
	ErrInvalidUsername       = errors.New("the username format is invalid")
	ErrUsernameAlreadyExists = errors.New("this username is already taken")
	ErrCreateUser            = errors.New("unable to create user account, please try again")
	ErrUpdateUser            = errors.New("unable to update user information, please try again")
	ErrDeleteUser            = errors.New("unable to delete user account, please try again")
	ErrUpdateUserStatus      = errors.New("unable to update user status, please try again")
	ErrGetAllUsers           = errors.New("unable to retrieve user list, please try again")
	ErrGetUserInformation    = errors.New("unable to retrieve user information")
	ErrActivateUser          = errors.New("unable to activate user account, please try again")
	ErrUpdateUserPassword    = errors.New("unable to update password, please try again")
)
