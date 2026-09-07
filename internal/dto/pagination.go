package dto

import "math"

type PaginationRequest struct {
	Page    int    `form:"page"     binding:"omitempty,min=1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Search  string `form:"search"   binding:"omitempty"`
	SortBy  string `form:"sort_by"  binding:"omitempty"`
	Sort    string `form:"sort"     binding:"omitempty,oneof=asc desc"`
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

func NormalizePageAndPerPage(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}

	if perPage < 5 {
		perPage = 5
	}

	return page, perPage
}

func Offset(page, perPage int) int {
	return (page - 1) * perPage
}
