package http

import (
	"net/http"

	"go-gin-clean/internal/application/usecase"
	"go-gin-clean/internal/delivery/http/response"
	"go-gin-clean/internal/dto"
	pkgerrors "go-gin-clean/pkg/errors"
	"go-gin-clean/pkg/message"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUseCase usecase.UserUseCase
}

func NewUserHandler(userUseCase usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.userUseCase.Login(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	response.Success(c, message.LoginSuccess, gin.H{
		"access_token": result.AccessToken,
	}, http.StatusOK)
}

func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.userUseCase.Register(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.RegisterSuccess, nil, http.StatusCreated)
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		response.Error(c, pkgerrors.WrapAppError(pkgerrors.Unauthorized, message.ErrTokenNotFound, err))
		return
	}

	result, err := h.userUseCase.RefreshToken(c.Request.Context(), cookie)
	if err != nil {
		response.Error(c, err)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	response.Success(c, message.RefreshSuccess, result.AccessToken, http.StatusOK)
}

func (h *UserHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrUnauthorized))
		return
	}

	if err := h.userUseCase.Logout(c.Request.Context(), userID.(string)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.LogoutSuccess, nil, http.StatusOK)
}

func (h *UserHandler) SendVerifyEmail(c *gin.Context) {
	var req dto.SendVerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.userUseCase.SendVerifyEmail(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.SendVerificationEmailSuccess, nil, http.StatusOK)
}

func (h *UserHandler) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.userUseCase.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.VerifyEmailSuccess, nil, http.StatusOK)
}

func (h *UserHandler) SendResetPassword(c *gin.Context) {
	var req dto.SendResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.userUseCase.SendResetPassword(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.SendResetPasswordEmailSuccess, nil, http.StatusOK)
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.userUseCase.ResetPassword(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.ResetPasswordSuccess, nil, http.StatusOK)
}

func (h *UserHandler) Profile(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrUnauthorized))
		return
	}

	result, err := h.userUseCase.GetUserByID(c.Request.Context(), userID.(string))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.GetUserInfoSuccess, result, http.StatusOK)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrUnauthorized))
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBind(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.userUseCase.UpdateUser(c.Request.Context(), userID.(string), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.UpdateUserSuccess, result, http.StatusOK)
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	var req dto.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PerPage <= 0 {
		req.PerPage = 10
	}

	result, err := h.userUseCase.GetAllUsers(c.Request.Context(), req.Page, req.PerPage, req.Search)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessPagination(c, message.GetAllUsersSuccess, result.Data, response.SetMeta(req.Page, req.PerPage, result.Total, result.TotalPages))
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")

	result, err := h.userUseCase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.GetUserInfoSuccess, result, http.StatusOK)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.userUseCase.CreateUser(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.CreateUserSuccess, result, http.StatusCreated)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	var req dto.UpdateUserRequest
	if err := c.ShouldBind(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.userUseCase.UpdateUser(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.UpdateUserSuccess, result, http.StatusOK)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrUnauthorized))
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.userUseCase.ChangePassword(c.Request.Context(), userID.(string), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.ChangePasswordSuccess, nil, http.StatusOK)
}

func (h *UserHandler) ChangeStatus(c *gin.Context) {
	userID := c.Param("id")

	var req dto.ChangeUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.userUseCase.ChangeStatus(c.Request.Context(), userID, &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.ChangeUserStatusSuccess, nil, http.StatusOK)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	if err := h.userUseCase.DeleteUser(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.DeleteUserSuccess, nil, http.StatusOK)
}
