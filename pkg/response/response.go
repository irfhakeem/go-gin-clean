package response

import (
	"errors"
	"net/http"
	"strings"

	pkgerror "go-gin-clean/pkg/error"
	"go-gin-clean/pkg/response/message"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status  bool              `json:"status"`
	Code    string            `json:"code,omitempty"`
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

func Success(c *gin.Context, code pkgerror.ErrCode, data any, httpCode int) {
	lang := detectLanguage(c)
	c.JSON(httpCode, Response{
		Status:  true,
		Message: message.Get(lang, code),
		Data:    data,
	})
}

func SuccessPagination(c *gin.Context, code pkgerror.ErrCode, data any, meta Meta) {
	lang := detectLanguage(c)
	c.JSON(http.StatusOK, Response{
		Status:  true,
		Message: message.Get(lang, code),
		Data:    data,
		Meta:    &meta,
	})
}

func Error(c *gin.Context, err error) {
	lang := detectLanguage(c)

	var appErr *pkgerror.AppError
	if !errors.As(err, &appErr) {
		appErr = pkgerror.InternalServerError(pkgerror.ErrInternalServerError, err)
	}

	c.JSON(appErr.HTTPStatus, Response{
		Status:  false,
		Code:    string(appErr.Code),
		Message: message.Get(lang, appErr.Code),
	})
}

func Reason(c *gin.Context, err error) string {
	lang := detectLanguage(c)

	var appErr *pkgerror.AppError
	if !errors.As(err, &appErr) {
		return err.Error()
	}
	return message.Get(lang, appErr.Code)
}

func ValidationError(c *gin.Context, errs map[string]string) {
	lang := detectLanguage(c)
	c.JSON(http.StatusUnprocessableEntity, Response{
		Status:  false,
		Code:    string(pkgerror.ErrValidationFailed),
		Message: message.Get(lang, pkgerror.ErrValidationFailed),
		Errors:  errs,
	})
}
