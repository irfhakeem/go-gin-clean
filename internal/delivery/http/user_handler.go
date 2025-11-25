package http

import (
	"go-gin-clean/internal/delivery/http/response"
	"go-gin-clean/internal/model"
	"go-gin-clean/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUseCase *usecase.UserUseCase
}

func NewUserHandler(userUseCase *usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "Failed to bind body", err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.userUseCase.Login(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, "Login failed", err.Error(), http.StatusUnauthorized)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	response.Success(c, "Login successful", gin.H{
		"access_token": result.AccessToken,
	}, http.StatusOK)
}

func (h *UserHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, "Failed to bind body", err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.userUseCase.Register(c.Request.Context(), &req); err != nil {
		response.Error(c, "Registration failed", err.Error(), http.StatusBadRequest)
		return
	}

	response.Success(c, "Registration successful", nil, http.StatusCreated)
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		response.Error(c, "Token not found", "Missing user credentials", http.StatusBadRequest)
		return
	}

	result, err := h.userUseCase.RefreshToken(c.Request.Context(), cookie)
	if err != nil {
		response.Error(c, "Failed to refresh token", err.Error(), http.StatusUnauthorized)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	response.Success(c, "Token refreshed successfully", result.AccessToken, http.StatusOK)
}

func (h *UserHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_pkid")
	if !exists {
		response.Error(c, "Unauthorized", "", http.StatusUnauthorized)
		return
	}

	err := h.userUseCase.Logout(c.Request.Context(), userID.(int64))
	if err != nil {
		response.Error(c, "Logout failed", err.Error(), http.StatusInternalServerError)
		return
	}
	response.Success(c, "Logout successful", nil, http.StatusOK)
}

func (h *UserHandler) SendVerifyEmail(c *gin.Context) {
	var req model.SendVerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "Failed to bind body", err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.userUseCase.SendVerifyEmail(c.Request.Context(), req); err != nil {
		response.Error(c, "Failed to send verification email", err.Error(), http.StatusInternalServerError)
		return
	}
	response.Success(c, "Verification email sent successfully", nil, http.StatusOK)
}

func (h *UserHandler) VerifyEmail(c *gin.Context) {
	var req model.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "Failed to bind body", err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.userUseCase.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		response.Error(c, "Email verification failed", err.Error(), http.StatusBadRequest)
		return
	}

	response.Success(c, "Email verified successfully", nil, http.StatusOK)
}

func (h *UserHandler) SendResetPassword(c *gin.Context) {
	var req model.SendResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "Failed to bind body", err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.userUseCase.SendResetPassword(c.Request.Context(), req); err != nil {
		response.Error(c, "Failed to send reset password email", err.Error(), http.StatusInternalServerError)
		return
	}

	response.Success(c, "Reset password email sent successfully", nil, http.StatusOK)
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req model.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "Failed to bind body", err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.userUseCase.ResetPassword(c.Request.Context(), &req); err != nil {
		response.Error(c, "Password reset failed", err.Error(), http.StatusBadRequest)
	}

	response.Success(c, "Password reset successful", nil, http.StatusOK)
}

func (h *UserHandler) Profile(c *gin.Context) {
	userCode, exist := c.Get("user_code")
	if !exist {
		response.Error(c, "Unauthorized", "user credentials not found", http.StatusUnauthorized)
		return
	}

	result, err := h.userUseCase.GetUserByCode(c.Request.Context(), userCode.(string))
	if err != nil {
		response.Error(c, "Failed to load profile", err.Error(), http.StatusInternalServerError)
		return
	}
	response.Success(c, "Profile loaded successfully", result, http.StatusOK)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userCode, exist := c.Get("user_code")
	if !exist {
		response.Error(c, "Unauthorized", "user credentials not found", http.StatusUnauthorized)
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, "Failed to bind body", err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.userUseCase.UpdateUser(c.Request.Context(), userCode.(string), &req)
	if err != nil {
		response.Error(c, "Failed to update profile", err.Error(), http.StatusBadRequest)
		return
	}
	response.Success(c, "Profile updated successfully", result, http.StatusOK)
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	var req model.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, "Failed to bind query", err.Error(), http.StatusBadRequest)
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
		response.Error(c, "Failed to get all users", err.Error(), http.StatusInternalServerError)
		return
	}
	response.SuccessPagination(c, result.Data, response.SetMeta(req.Page, req.PerPage, result.Total, result.TotalPages))
}

func (h *UserHandler) GetUserByCode(c *gin.Context) {
	userCode := c.Param("code")

	result, err := h.userUseCase.GetUserByCode(c.Request.Context(), userCode)
	if err != nil {
		response.Error(c, "User not found", err.Error(), http.StatusNotFound)
		return
	}
	response.Success(c, "User retrieved successfully", result, http.StatusOK)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "Failed to bind body", err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.userUseCase.CreateUser(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, "Failed to create user", err.Error(), http.StatusBadRequest)
		return
	}
	response.Success(c, "User created successfully", result, http.StatusCreated)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userCode := c.Param("code")

	var req model.UpdateUserRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, "Failed to bind body", err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.userUseCase.UpdateUser(c.Request.Context(), userCode, &req)
	if err != nil {
		response.Error(c, "Failed to update user", err.Error(), http.StatusBadRequest)
		return
	}
	response.Success(c, "User updated successfully", result, http.StatusOK)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userPKID, exist := c.Get("user_pkid")
	if !exist {
		response.Error(c, "Unauthorized", "user credentials not found", http.StatusUnauthorized)
		return
	}

	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "Failed to bind body", err.Error(), http.StatusBadRequest)
		return
	}

	err := h.userUseCase.ChangePassword(c.Request.Context(), userPKID.(int64), &req)
	if err != nil {
		response.Error(c, "Password change failed", err.Error(), http.StatusBadRequest)
		return
	}
	response.Success(c, "Password changed successfully", nil, http.StatusOK)
}

func (h *UserHandler) ChangeStatus(c *gin.Context) {
	userCode := c.Param("code")

	var req model.ChangeUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "Failed to bind body", err.Error(), http.StatusBadRequest)
		return
	}

	err := h.userUseCase.ChangeStatus(c.Request.Context(), userCode, req)
	if err != nil {
		response.Error(c, "Failed to change user status", err.Error(), http.StatusInternalServerError)
		return
	}
	response.Success(c, "User status changed successfully", nil, http.StatusOK)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userCode := c.Param("code")

	if err := h.userUseCase.DeleteUser(c.Request.Context(), userCode); err != nil {
		response.Error(c, "Failed to delete user", err.Error(), http.StatusInternalServerError)
		return
	}

	response.Success(c, "User deleted successfully", nil, http.StatusOK)
}
