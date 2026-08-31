package response

import (
	"errors"
	"net/http"
	"strings"

	"go-gin-clean/internal/dto/validator"
	pkgerrors "go-gin-clean/pkg/errors"
	"go-gin-clean/pkg/message"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status  bool              `json:"status"`
	Key     string            `json:"code,omitempty"`
	Message string            `json:"message"`
	Data    any               `json:"data,omitempty"`
	Errors  map[string]string `json:"errors,omitempty"`
	Meta    *Meta             `json:"meta,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func detectLanguage(c *gin.Context) message.Language {
	header := strings.ToLower(c.GetHeader("Accept-Language"))
	switch {
	case strings.HasPrefix(header, "en"):
		return message.EN
	case strings.HasPrefix(header, "id"):
		return message.ID
	default:
		return message.EN
	}
}

func SetMeta(page, perPage, total, totalPages int) Meta {
	return Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

func Success(c *gin.Context, key message.MessageKey, data any, httpKey int) {
	lang := detectLanguage(c)
	c.JSON(httpKey, Response{
		Status:  true,
		Message: message.Get(lang, key),
		Data:    data,
	})
}

func SuccessPagination(c *gin.Context, key message.MessageKey, data any, meta Meta) {
	lang := detectLanguage(c)
	c.JSON(http.StatusOK, Response{
		Status:  true,
		Message: message.Get(lang, key),
		Data:    data,
		Meta:    &meta,
	})
}

func Error(c *gin.Context, err error) {
	lang := detectLanguage(c)

	var appErr *pkgerrors.AppError
	if !errors.As(err, &appErr) {
		c.JSON(http.StatusInternalServerError, Response{
			Status:  false,
			Key:     message.ErrInternalServerError.String(),
			Message: message.Get(lang, message.ErrInternalServerError),
		})
		return
	}

	var status int
	switch appErr.Type {
	case pkgerrors.BadRequest:
		status = http.StatusBadRequest
	case pkgerrors.Unauthorized:
		status = http.StatusUnauthorized
	case pkgerrors.Forbidden:
		status = http.StatusForbidden
	case pkgerrors.NotFound:
		status = http.StatusNotFound
	case pkgerrors.Conflict:
		status = http.StatusConflict
	case pkgerrors.Unprocessable:
		status = http.StatusUnprocessableEntity
	default:
		status = http.StatusInternalServerError
	}

	c.JSON(status, Response{
		Status:  false,
		Key:     appErr.Key.String(),
		Message: message.Get(lang, appErr.Key),
	})
}

func Reason(c *gin.Context, err error) string {
	lang := detectLanguage(c)

	var appErr *pkgerrors.AppError
	if !errors.As(err, &appErr) {
		return err.Error()
	}
	return message.Get(lang, appErr.Key)
}

func ValidationError(c *gin.Context, err error) {
	lang := detectLanguage(c)

	errs, ok := validator.BuildValidationErrors(err)

	if !ok {
		c.JSON(http.StatusUnprocessableEntity, Response{
			Status:  false,
			Key:     message.ErrValidationFailed.String(),
			Message: message.Get(lang, message.ErrValidationFailed),
		})
		return
	}

	c.JSON(http.StatusUnprocessableEntity, Response{
		Status:  false,
		Key:     message.ErrInvalidRequestBody.String(),
		Message: message.Get(lang, message.ErrInvalidRequestBody),
		Errors:  errs,
	})
}
