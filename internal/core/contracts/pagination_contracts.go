package contracts

type PaginationRequest struct {
	Page    int
	PerPage int
	Search  string
}

type PaginationResponse[T any] struct {
	Data       []T
	Page       int
	PerPage    int
	Total      int
	TotalPages int
}

func NewPaginationResponse[T any](data []T, page, perPage, total int) *PaginationResponse[T] {
	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	return &PaginationResponse[T]{
		Data:       data,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

func Offset(page, pageSize int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * pageSize
}
