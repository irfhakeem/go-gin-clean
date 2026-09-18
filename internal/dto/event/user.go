package event

import "github.com/google/uuid"

type (
	UserEvent struct {
		UserID uuid.UUID `json:"user_id"`
		Name   string    `json:"name"`
	}

	UserRegisterEvent struct {
		UserEvent
		Email           string `json:"email"`
		VerificationURL string `json:"verification_url"`
	}

	UserResetPasswordEvent struct {
		UserEvent
		Email    string `json:"email"`
		ResetURL string `json:"reset_url"`
	}
)
