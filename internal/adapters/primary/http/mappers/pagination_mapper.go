package mappers

import (
	"go-gin-clean/internal/adapters/primary/http/dto"
	"go-gin-clean/internal/core/contracts"
)

// paginationMapper implements the PaginationMapper interface
type paginationMapper struct{}

// NewPaginationMapper creates a new pagination mapper
func NewPaginationMapper() PaginationMapper {
	return &paginationMapper{}
}

func (m *paginationMapper) RequestToContract(req *dto.PaginationRequest) *contracts.PaginationRequest {
	return &contracts.PaginationRequest{
		Page:    req.Page,
		PerPage: req.PerPage,
		Search:  req.Search,
	}
}

func (m *paginationMapper) UserInfoResponseToDTO(resp *contracts.PaginationResponse[contracts.UserInfo]) *dto.PaginationResponse[dto.UserInfo] {
	userMapper := NewUserMapper()
	dtoUsers := make([]dto.UserInfo, len(resp.Data))
	for i, user := range resp.Data {
		dtoUsers[i] = *userMapper.UserInfoToDTO(&user)
	}

	return &dto.PaginationResponse[dto.UserInfo]{
		Data:       dtoUsers,
		Page:       resp.Page,
		PerPage:    resp.PerPage,
		Total:      resp.Total,
		TotalPages: resp.TotalPages,
	}
}
