package mq

import (
	"context"
	"encoding/json"
	"go-gin-clean/internal/dto/event"
	"go-gin-clean/internal/usecase"
)

type EmailEventHandler struct {
	emailUsecase usecase.EmailUseCaseInterface
}

func NewEmailEventHandler(emailUsecase usecase.EmailUseCaseInterface) *EmailEventHandler {
	return &EmailEventHandler{
		emailUsecase: emailUsecase,
	}
}

func (h *EmailEventHandler) HandleUserVerifyEmail(ctx context.Context, payload []byte) error {
	var data event.RegisterEvent

	if err := json.Unmarshal(payload, &data); err != nil {
		return NonRetryable(err)
	}

	return h.emailUsecase.SendVerifyEmail(data.Email, data.Name, data.VerificationURL)
}

func (h *EmailEventHandler) HandleUserResetPasswordEmail(ctx context.Context, payload []byte) error {
	var data event.ResetPasswordEvent

	if err := json.Unmarshal(payload, &data); err != nil {
		return NonRetryable(err)
	}

	return h.emailUsecase.SendResetPasswordEmail(data.Email, data.Name, data.ResetURL)
}
