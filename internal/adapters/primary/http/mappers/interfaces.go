package mappers

import (
	"go-gin-clean/internal/adapters/primary/http/dto"
	"go-gin-clean/internal/core/contracts"
)

type UserMapper interface {
	// DTO to Contract mappings
	LoginRequestToContract(req *dto.LoginRequest) *contracts.LoginRequest
	RegisterRequestToContract(req *dto.RegisterRequest) *contracts.RegisterRequest
	ResetPasswordRequestToContract(req *dto.ResetPasswordRequest) *contracts.ResetPasswordRequest
	ChangePasswordRequestToContract(req *dto.ChangePasswordRequest) *contracts.ChangePasswordRequest
	CreateUserRequestToContract(req *dto.CreateUserRequest) *contracts.CreateUserRequest
	UpdateUserRequestToContract(req *dto.UpdateUserRequest) *contracts.UpdateUserRequest
	PaginationRequestToContract(req *dto.PaginationRequest) *contracts.PaginationRequest

	// Contract to DTO mappings
	LoginResponseToDTO(resp *contracts.LoginResponse) *dto.LoginResponse
	RefreshTokenResponseToDTO(resp *contracts.RefreshTokenResponse) *dto.RefreshTokenResponse
	UserInfoToDTO(user *contracts.UserInfo) *dto.UserInfo
	PaginationResponseToDTO(resp *contracts.PaginationResponse[contracts.UserInfo]) *dto.PaginationResponse[dto.UserInfo]
}

type PaginationMapper interface {
	RequestToContract(req *dto.PaginationRequest) *contracts.PaginationRequest
	UserInfoResponseToDTO(resp *contracts.PaginationResponse[contracts.UserInfo]) *dto.PaginationResponse[dto.UserInfo]
}
