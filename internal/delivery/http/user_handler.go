package http

import (
	"net/http"

	"go-gin-clean/internal/dto"
	"go-gin-clean/internal/dto/validator"
	"go-gin-clean/internal/usecase"
	pkgerror "go-gin-clean/pkg/error"
	"go-gin-clean/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUseCase usecase.UserUseCaseInterface
}

func NewUserHandler(userUseCase usecase.UserUseCaseInterface) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

func bindError(c *gin.Context, err error) {
	if errs, ok := validator.BuildValidationErrors(err); ok {
		response.ValidationError(c, errs)
	} else {
		response.Error(c, pkgerror.BadRequest(pkgerror.ErrInvalidRequestBody, err))
	}
}

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
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

	response.Success(c, pkgerror.LoginSuccess, gin.H{
		"access_token": result.AccessToken,
	}, http.StatusOK)
}

func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}

	if err := h.userUseCase.Register(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, pkgerror.RegisterSuccess, nil, http.StatusCreated)
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		response.Error(c, pkgerror.BadRequest(pkgerror.ErrTokenNotFound, err))
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

	response.Success(c, pkgerror.RefreshSuccess, result.AccessToken, http.StatusOK)
}

func (h *UserHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, pkgerror.Unauthorized(pkgerror.ErrUnauthorized, nil))
		return
	}

	if err := h.userUseCase.Logout(c.Request.Context(), userID.(string)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, pkgerror.LogoutSuccess, nil, http.StatusOK)
}

func (h *UserHandler) SendVerifyEmail(c *gin.Context) {
	var req dto.SendVerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}

	if err := h.userUseCase.SendVerifyEmail(c.Request.Context(), req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, pkgerror.SendVerificationEmailSuccess, nil, http.StatusOK)
}

func (h *UserHandler) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}

	if err := h.userUseCase.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, pkgerror.VerifyEmailSuccess, nil, http.StatusOK)
}

func (h *UserHandler) SendResetPassword(c *gin.Context) {
	var req dto.SendResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}

	if err := h.userUseCase.SendResetPassword(c.Request.Context(), req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, pkgerror.SendResetPasswordEmailSuccess, nil, http.StatusOK)
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}

	if err := h.userUseCase.ResetPassword(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, pkgerror.ResetPasswordSuccess, nil, http.StatusOK)
}

func (h *UserHandler) Profile(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		response.Error(c, pkgerror.Unauthorized(pkgerror.ErrUnauthorized, nil))
		return
	}

	result, err := h.userUseCase.GetUserByID(c.Request.Context(), userID.(string))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, pkgerror.GetUserInfoSuccess, result, http.StatusOK)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		response.Error(c, pkgerror.Unauthorized(pkgerror.ErrUnauthorized, nil))
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBind(&req); err != nil {
		bindError(c, err)
		return
	}

	result, err := h.userUseCase.UpdateUser(c.Request.Context(), userID.(string), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, pkgerror.UpdateUserSuccess, result, http.StatusOK)
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	var req dto.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		bindError(c, err)
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

	response.SuccessPagination(c, pkgerror.GetAllUsersSuccess, result.Data, response.SetMeta(req.Page, req.PerPage, result.Total, result.TotalPages))
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	userID := c.Param("id")

	result, err := h.userUseCase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, pkgerror.GetUserInfoSuccess, result, http.StatusOK)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}

	result, err := h.userUseCase.CreateUser(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, pkgerror.CreateUserSuccess, result, http.StatusCreated)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	var req dto.UpdateUserRequest
	if err := c.ShouldBind(&req); err != nil {
		bindError(c, err)
		return
	}

	result, err := h.userUseCase.UpdateUser(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, pkgerror.UpdateUserSuccess, result, http.StatusOK)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		response.Error(c, pkgerror.Unauthorized(pkgerror.ErrUnauthorized, nil))
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}

	if err := h.userUseCase.ChangePassword(c.Request.Context(), userID.(string), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, pkgerror.ChangePasswordSuccess, nil, http.StatusOK)
}

func (h *UserHandler) ChangeStatus(c *gin.Context) {
	userID := c.Param("id")

	var req dto.ChangeUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
		return
	}

	if err := h.userUseCase.ChangeStatus(c.Request.Context(), userID, req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, pkgerror.ChangeUserStatusSuccess, nil, http.StatusOK)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	if err := h.userUseCase.DeleteUser(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, pkgerror.DeleteUserSuccess, nil, http.StatusOK)
}
