package model

import "math"

type PaginationRequest struct {
	Page    int    `form:"page" json:"page"`
	PerPage int    `form:"per_page" json:"per_page"`
	Search  string `form:"search,omitempty" json:"search,omitempty"`
}

type PaginationResponse[T any] struct {
	Data       []T `json:"data"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func NewPaginationResponse[T any](data []T, page, perPage, total int) *PaginationResponse[T] {
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	return &PaginationResponse[T]{
		Data:       data,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

func Offset(page, perPage int) int {
	if page < 1 {
		page = 1
	}

	if perPage < 5 {
		perPage = 5
	}

	return (page - 1) * perPage
}
