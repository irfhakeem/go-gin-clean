package mappers

import (
	"bytes"
	"io"

	"go-gin-clean/internal/adapters/primary/http/dto"
	"go-gin-clean/internal/core/contracts"
)

// userMapper implements the UserMapper interface
type userMapper struct{}

// NewUserMapper creates a new user mapper
func NewUserMapper() UserMapper {
	return &userMapper{}
}

// DTO to Contract mappings
func (m *userMapper) LoginRequestToContract(req *dto.LoginRequest) *contracts.LoginRequest {
	return &contracts.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}
}

func (m *userMapper) RegisterRequestToContract(req *dto.RegisterRequest) *contracts.RegisterRequest {
	return &contracts.RegisterRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}
}

func (m *userMapper) ResetPasswordRequestToContract(req *dto.ResetPasswordRequest) *contracts.ResetPasswordRequest {
	return &contracts.ResetPasswordRequest{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	}
}

func (m *userMapper) ChangePasswordRequestToContract(req *dto.ChangePasswordRequest) *contracts.ChangePasswordRequest {
	return &contracts.ChangePasswordRequest{
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}
}

func (m *userMapper) CreateUserRequestToContract(req *dto.CreateUserRequest) *contracts.CreateUserRequest {
	return &contracts.CreateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Gender:   req.Gender,
	}
}

func (m *userMapper) UpdateUserRequestToContract(req *dto.UpdateUserRequest) *contracts.UpdateUserRequest {
	contractReq := &contracts.UpdateUserRequest{
		Name:   req.Name,
		Gender: req.Gender,
	}

	// Convert multipart.FileHeader to contracts.FileUpload
	if req.Avatar != nil {
		file, err := req.Avatar.Open()
		if err == nil {
			defer file.Close()

			// Read file content into buffer
			buf := new(bytes.Buffer)
			io.Copy(buf, file)

			contractReq.Avatar = &contracts.FileUpload{
				Filename: req.Avatar.Filename,
				Size:     req.Avatar.Size,
				Content:  bytes.NewReader(buf.Bytes()),
			}
		}
	}

	return contractReq
}

func (m *userMapper) PaginationRequestToContract(req *dto.PaginationRequest) *contracts.PaginationRequest {
	return &contracts.PaginationRequest{
		Page:    req.Page,
		PerPage: req.PerPage,
		Search:  req.Search,
	}
}

// Contract to DTO mappings
func (m *userMapper) LoginResponseToDTO(resp *contracts.LoginResponse) *dto.LoginResponse {
	return &dto.LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		User:         *m.UserInfoToDTO(&resp.User),
	}
}

func (m *userMapper) RefreshTokenResponseToDTO(resp *contracts.RefreshTokenResponse) *dto.RefreshTokenResponse {
	return &dto.RefreshTokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}
}

func (m *userMapper) UserInfoToDTO(user *contracts.UserInfo) *dto.UserInfo {
	return &dto.UserInfo{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Gender:   user.Gender,
		IsActive: user.IsActive,
	}
}

func (m *userMapper) PaginationResponseToDTO(resp *contracts.PaginationResponse[contracts.UserInfo]) *dto.PaginationResponse[dto.UserInfo] {
	dtoUsers := make([]dto.UserInfo, len(resp.Data))
	for i, user := range resp.Data {
		dtoUsers[i] = *m.UserInfoToDTO(&user)
	}

	return &dto.PaginationResponse[dto.UserInfo]{
		Data:       dtoUsers,
		Page:       resp.Page,
		PerPage:    resp.PerPage,
		Total:      resp.Total,
		TotalPages: resp.TotalPages,
	}
}
