package model

import "github.com/google/uuid"

type (
	UserEvent struct {
		UserID uuid.UUID `json:"user_id"`
		Name   string    `json:"name"`
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
