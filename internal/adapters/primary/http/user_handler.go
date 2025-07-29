package http

import (
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

type Response struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

type Meta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status:  false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	result, err := h.userUseCase.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, Response{
			Status:  false,
			Message: "Login failed",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  true,
		Message: "Login successful",
		Data:    result,
	})
}

func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status:  false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	err := h.userUseCase.Register(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status:  false,
			Message: "Registration failed",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Status:  true,
		Message: "Registration successful. Please verify your email.",
	})
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	refreshToken := c.GetHeader("X-Refresh-Token")
	if refreshToken == "" {
		c.JSON(http.StatusBadRequest, Response{
			Status:  false,
			Message: "Refresh token required",
			Error:   "Missing refresh token",
		})
		return
	}

	result, err := h.userUseCase.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, Response{
			Status:  false,
			Message: "Token refresh failed",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  true,
		Message: "Token refreshed successfully",
		Data:    result,
	})
}

func (h *UserHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, Response{
			Status:  false,
			Message: "Unauthorized",
			Error:   "User ID not found in context",
		})
		return
	}

	err := h.userUseCase.Logout(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status:  false,
			Message: "Logout failed",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  true,
		Message: "Logout successful",
	})
}

func (h *UserHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, Response{
			Status:  false,
			Message: "Token required",
			Error:   "Missing verification token",
		})
		return
	}

	err := h.userUseCase.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status:  false,
			Message: "Email verification failed",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  true,
		Message: "Email verified successfully",
	})
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	var req dto.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status:  false,
			Message: "Invalid query parameters",
			Error:   err.Error(),
		})
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
		c.JSON(http.StatusInternalServerError, Response{
			Status:  false,
			Message: "Failed to get users",
			Error:   err.Error(),
		})
		return
	}

	meta := Meta{
		Page:       result.Page,
		PerPage:    result.PerPage,
		Total:      result.Total,
		TotalPages: result.TotalPages,
	}

	c.JSON(http.StatusOK, Response{
		Status:  true,
		Message: "Users retrieved successfully",
		Data:    result.Data,
		Meta:    &meta,
	})
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status:  false,
			Message: "Invalid user ID",
			Error:   err.Error(),
		})
		return
	}

	result, err := h.userUseCase.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Status:  false,
			Message: "User not found",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  true,
		Message: "User retrieved successfully",
		Data:    result,
	})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status:  false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	result, err := h.userUseCase.CreateUser(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status:  false,
			Message: "Failed to create user",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Status:  true,
		Message: "User created successfully",
		Data:    result,
	})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status:  false,
			Message: "Invalid user ID",
			Error:   err.Error(),
		})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status:  false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	result, err := h.userUseCase.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status:  false,
			Message: "Failed to update user",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  true,
		Message: "User updated successfully",
		Data:    result,
	})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, Response{
			Status:  false,
			Message: "Unauthorized",
			Error:   "User ID not found in context",
		})
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status:  false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	err := h.userUseCase.ChangePassword(c.Request.Context(), userID.(int64), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status:  false,
			Message: "Failed to change password",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  true,
		Message: "Password changed successfully",
	})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status:  false,
			Message: "Invalid user ID",
			Error:   err.Error(),
		})
		return
	}

	err = h.userUseCase.DeleteUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status:  false,
			Message: "Failed to delete user",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  true,
		Message: "User deleted successfully",
	})
}
