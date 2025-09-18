package handlers

import (
	"go-gin-clean/internal/adapters/primary/http/messages"
	"go-gin-clean/internal/adapters/primary/http/response"
	"go-gin-clean/internal/core/dto"
	"go-gin-clean/internal/core/ports"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userUseCase ports.UserUseCase
}

func NewUserHandler(userUseCase ports.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, messages.FAILED_TO_BIND_BODY, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.userUseCase.Login(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, messages.FAILED_LOGIN, err.Error(), http.StatusUnauthorized)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	response.Success(c, messages.SUCCESS_LOGIN, gin.H{
		"access_token": result.AccessToken,
		"user":         result.User,
	}, http.StatusOK)
}

func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, messages.FAILED_TO_BIND_BODY, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.userUseCase.Register(c.Request.Context(), &req); err != nil {
		response.Error(c, messages.FAILED_REGISTRATION, err.Error(), http.StatusBadRequest)
		return
	}

	response.Success(c, messages.SUCCESS_REGISTRATION, nil, http.StatusCreated)
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		response.Error(c, messages.FAILED_TOKEN_NOT_FOUND, "Missing user credentials", http.StatusBadRequest)
		return
	}

	result, err := h.userUseCase.RefreshToken(c.Request.Context(), cookie)
	if err != nil {
		response.Error(c, messages.FAILED_REFRESH_TOKEN, err.Error(), http.StatusUnauthorized)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	response.Success(c, messages.SUCCESS_REFRESH_TOKEN, result.AccessToken, http.StatusOK)
}

func (h *UserHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, messages.FAILED_UNAUTHORIZED, "", http.StatusUnauthorized)
		return
	}

	err := h.userUseCase.Logout(c.Request.Context(), userID.(int64))
	if err != nil {
		response.Error(c, messages.FAILED_LOGOUT, err.Error(), http.StatusInternalServerError)
		return
	}
	response.Success(c, messages.SUCCESS_LOGOUT, nil, http.StatusOK)
}

func (h *UserHandler) SendVerifyEmail(c *gin.Context) {
	var req dto.SendVerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, messages.FAILED_TO_BIND_BODY, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.userUseCase.SendVerifyEmail(c.Request.Context(), req.Email); err != nil {
		response.Error(c, messages.FAILED_SEND_EMAIL_VERIFY, err.Error(), http.StatusInternalServerError)
		return
	}
	response.Success(c, messages.SUCCESS_SEND_EMAIL_VERIFY, nil, http.StatusOK)
}

func (h *UserHandler) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, messages.FAILED_TO_BIND_BODY, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.userUseCase.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		response.Error(c, messages.FAILED_VERIFY_EMAIL, err.Error(), http.StatusBadRequest)
		return
	}

	response.Success(c, messages.SUCCESS_VERIFY_EMAIL, nil, http.StatusOK)
}

func (h *UserHandler) SendResetPassword(c *gin.Context) {
	var req dto.SendResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, messages.FAILED_TO_BIND_BODY, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.userUseCase.SendResetPassword(c.Request.Context(), req.Email); err != nil {
		response.Error(c, messages.FAILED_SEND_EMAIL_RESET_PASSWORD, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Success(c, messages.SUCCESS_SEND_EMAIL_RESET_PASSWORD, nil, http.StatusOK)
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, messages.FAILED_TO_BIND_BODY, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.userUseCase.ResetPassword(c.Request.Context(), &req); err != nil {
		response.Error(c, messages.FAILED_RESET_PASSWORD, err.Error(), http.StatusBadRequest)
	}

	response.Success(c, messages.SUCCESS_RESET_PASSWORD, nil, http.StatusOK)
}

func (h *UserHandler) Profile(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		response.Error(c, messages.FAILED_UNAUTHORIZED, "user credentials not found", http.StatusUnauthorized)
		return
	}

	result, err := h.userUseCase.GetUserByID(c.Request.Context(), userID.(int64))
	if err != nil {
		response.Error(c, messages.FAILED_LOAD_PROFILE, err.Error(), http.StatusInternalServerError)
		return
	}
	response.Success(c, messages.SUCCESS_LOAD_PROFILE, result, http.StatusOK)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		response.Error(c, messages.FAILED_UNAUTHORIZED, "user credentials not found", http.StatusUnauthorized)
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, messages.FAILED_TO_BIND_BODY, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.userUseCase.UpdateUser(c.Request.Context(), userID.(int64), &req)
	if err != nil {
		response.Error(c, messages.FAILED_UPDATE_PROFILE, err.Error(), http.StatusBadRequest)
		return
	}
	response.Success(c, messages.SUCCESS_UPDATE_PROFILE, result, http.StatusOK)
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	var req dto.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, messages.FAILED_TO_BIND_QUERY, err.Error(), http.StatusBadRequest)
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
		response.Error(c, messages.FAILED_GET_ALL_USERS, err.Error(), http.StatusInternalServerError)
		return
	}
	response.SuccessPagination(c, result.Data, response.SetMeta(req.Page, req.PerPage, result.Total, result.TotalPages))
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.Error(c, messages.FAILED_TO_BIND_PARAMS, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.userUseCase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, messages.FAILED_USER_NOT_FOUND, err.Error(), http.StatusNotFound)
		return
	}
	response.Success(c, messages.SUCCESS_GET_USER, result, http.StatusOK)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, messages.FAILED_TO_BIND_BODY, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.userUseCase.CreateUser(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, messages.FAILED_CREATE_USER, err.Error(), http.StatusBadRequest)
		return
	}
	response.Success(c, messages.SUCCESS_CREATE_USER, result, http.StatusCreated)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.Error(c, messages.FAILED_TO_BIND_PARAMS, err.Error(), http.StatusBadRequest)
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBind(&req); err != nil {
		response.Error(c, messages.FAILED_TO_BIND_BODY, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.userUseCase.UpdateUser(c.Request.Context(), userID, &req)
	if err != nil {
		response.Error(c, messages.FAILED_UPDATE_USER, err.Error(), http.StatusBadRequest)
		return
	}
	response.Success(c, messages.SUCCESS_UPDATE_USER, result, http.StatusOK)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, messages.FAILED_UNAUTHORIZED, "user credentials not found", http.StatusUnauthorized)
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, messages.FAILED_TO_BIND_BODY, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.userUseCase.ChangePassword(c.Request.Context(), userID.(int64), &req)
	if err != nil {
		response.Error(c, messages.FAILED_PASSWORD_CHANGE, err.Error(), http.StatusBadRequest)
		return
	}

	response.Success(c, messages.SUCCESS_PASSWORD_CHANGE, nil, http.StatusOK)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, messages.FAILED_TO_BIND_PARAMS, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.userUseCase.DeleteUser(c.Request.Context(), id)
	if err != nil {
		response.Error(c, messages.FAILED_DELETE_USER, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Success(c, messages.SUCCESS_DELETE_USER, nil, http.StatusNoContent)
}
