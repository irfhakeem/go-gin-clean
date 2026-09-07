package http

import (
	"net/http"

	"go-gin-clean/internal/application/usecase"
	"go-gin-clean/internal/delivery/http/response"
	"go-gin-clean/internal/domain/policy"
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

func (h *UserHandler) Profile(c *gin.Context) {
	userID, existUser := c.Get("user_id")
	actor, existActor := c.Get("actor")
	if !existUser || !existActor {
		response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrUnauthorized))
		return
	}

	result, err := h.userUseCase.GetUserByID(c.Request.Context(), actor.(policy.Actor), userID.(string))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.GetUserInfoSuccess, result, http.StatusOK)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, existUser := c.Get("user_id")
	actor, existActor := c.Get("actor")
	if !existUser || !existActor {
		response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrUnauthorized))
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBind(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.userUseCase.UpdateUser(c.Request.Context(), actor.(policy.Actor), userID.(string), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.UpdateUserSuccess, result, http.StatusOK)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, existUser := c.Get("user_id")
	actor, existActor := c.Get("actor")
	if !existUser || !existActor {
		response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrUnauthorized))
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.userUseCase.ChangePassword(c.Request.Context(), actor.(policy.Actor), userID.(string), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.ChangePasswordSuccess, nil, http.StatusOK)
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	actor, exist := c.Get("actor")
	if !exist {
		response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrUnauthorized))
		return
	}

	var req dto.GetAllUserQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	req.Page, req.PerPage = dto.NormalizePageAndPerPage(req.Page, req.PerPage)
	offset := dto.Offset(req.Page, req.PerPage)

	result, err := h.userUseCase.GetAllUsers(c.Request.Context(), actor.(policy.Actor), &req, offset)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessPagination(c, message.GetAllUsersSuccess, result.Data, response.SetMeta(req.Page, req.PerPage, result.Total, result.TotalPages))
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	actor, exist := c.Get("actor")
	if !exist {
		response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrUnauthorized))
		return
	}

	userID := c.Param("id")

	result, err := h.userUseCase.GetUserByID(c.Request.Context(), actor.(policy.Actor), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.GetUserInfoSuccess, result, http.StatusOK)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	actor, exist := c.Get("actor")
	if !exist {
		response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrUnauthorized))
		return
	}

	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.userUseCase.CreateUser(c.Request.Context(), actor.(policy.Actor), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.CreateUserSuccess, result, http.StatusCreated)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	actor, exist := c.Get("actor")
	if !exist {
		response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrUnauthorized))
		return
	}

	userID := c.Param("id")

	var req dto.UpdateUserRequest
	if err := c.ShouldBind(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.userUseCase.UpdateUser(c.Request.Context(), actor.(policy.Actor), userID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.UpdateUserSuccess, result, http.StatusOK)
}

func (h *UserHandler) ChangeStatus(c *gin.Context) {
	actor, exist := c.Get("actor")
	if !exist {
		response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrUnauthorized))
		return
	}

	userID := c.Param("id")

	var req dto.ChangeUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.userUseCase.ChangeStatus(c.Request.Context(), actor.(policy.Actor), userID, &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.ChangeUserStatusSuccess, nil, http.StatusOK)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	actor, exist := c.Get("actor")
	if !exist {
		response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrUnauthorized))
		return
	}

	userID := c.Param("id")

	if err := h.userUseCase.DeleteUser(c.Request.Context(), actor.(policy.Actor), userID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.DeleteUserSuccess, nil, http.StatusOK)
}
