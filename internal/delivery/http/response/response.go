package response

import (
	"github.com/gin-gonic/gin"
)

type Response struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func SetMeta(page, perPage, total, totalPages int) Meta {
	return Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

func Success(c *gin.Context, message string, data any, code int) {
	c.JSON(code, Response{
		Status:  true,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, message string, err string, code int) {
	c.JSON(code, Response{
		Status:  false,
		Message: message,
		Error:   err,
	})
}

func SuccessPagination(c *gin.Context, data any, meta Meta) {
	c.JSON(200, Response{
		Status:  true,
		Message: "Data retrieved successfully",
		Data:    data,
		Meta:    &meta,
	})
}
