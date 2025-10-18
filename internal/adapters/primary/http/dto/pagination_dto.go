package dto

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
