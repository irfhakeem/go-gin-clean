package inbound

type EmailUseCase interface {
	SendVerifyEmail(to, name, url string) error
	SendResetPasswordEmail(to, name, url string) error
}
