package mq

import (
	"context"
	"encoding/json"
	"errors"
	"go-gin-clean/internal/application/usecase"
	"go-gin-clean/internal/domain/entity"
	pkgerrors "go-gin-clean/pkg/errors"
	"net/textproto"
)

type EmailEventHandler struct {
	emailUsecase usecase.EmailUseCase
}

func NewEmailEventHandler(emailUsecase usecase.EmailUseCase) *EmailEventHandler {
	return &EmailEventHandler{
		emailUsecase: emailUsecase,
	}
}

func (h *EmailEventHandler) HandleUserVerifyEmail(ctx context.Context, payload []byte) error {
	var data entity.UserRegisterEvent

	if err := checkJSON(payload, &data); err != nil {
		return err
	}

	err := h.emailUsecase.SendVerifyEmail(data.Email, data.Name, data.VerificationURL)
	return checkSMTPError(err)
}

func (h *EmailEventHandler) HandleUserResetPasswordEmail(ctx context.Context, payload []byte) error {
	var data entity.UserResetPasswordEvent

	if err := checkJSON(payload, &data); err != nil {
		return err
	}

	err := h.emailUsecase.SendResetPasswordEmail(data.Email, data.Name, data.ResetURL)
	return checkSMTPError(err)
}

func checkJSON(payload []byte, v any) error {
	if err := json.Unmarshal(payload, v); err != nil {
		return pkgerrors.NewConsumerError(
			pkgerrors.NonRetryable,
			err,
		)
	}
	return nil
}

func checkSMTPError(err error) error {
	var smtpErr *textproto.Error
	if errors.As(err, &smtpErr) {
		if smtpErr.Code >= 400 && smtpErr.Code < 500 {
			return pkgerrors.NewConsumerError(
				pkgerrors.Retryable,
				err,
			)
		}
	}

	return pkgerrors.NewConsumerError(
		pkgerrors.NonRetryable,
		err,
	)
}
