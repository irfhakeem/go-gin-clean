package model

type (
	UserEvent struct {
		UserPKID int64  `json:"user_pkid"`
		Name     string `json:"name"`
	}

	RegisterEvent struct {
		UserEvent
		Email           string `json:"email"`
		VerificationURL string `json:"verification_url"`
	}

	ResetPasswordEvent struct {
		UserEvent
		Email    string `json:"email"`
		ResetURL string `json:"reset_url"`
	}
)
