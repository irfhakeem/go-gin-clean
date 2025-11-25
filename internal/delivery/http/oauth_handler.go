package http

import (
	"go-gin-clean/internal/delivery/http/response"
	"go-gin-clean/internal/model"
	"go-gin-clean/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OAuthHandler struct {
	userUseCase *usecase.UserUseCase
}

func NewOAuthHandler(userUseCase *usecase.UserUseCase) *OAuthHandler {
	return &OAuthHandler{
		userUseCase: userUseCase,
	}
}

func (h *OAuthHandler) GetLoginURL(c *gin.Context) {
	var req model.OAuthLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "Failed to parse JSON", err.Error(), http.StatusBadRequest)
		return
	}

	urlResp, err := h.userUseCase.GetOAuthLoginURL(c.Request.Context(), req.Provider, req.AppID)
	if err != nil {
		response.Error(c, "Failed to generate OAuth URL", err.Error(), http.StatusInternalServerError)
		return
	}

	response.Success(c, "OAuth URL generated successfully", urlResp, http.StatusOK)
}

func (h *OAuthHandler) CallBack(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")
	appId := c.Query("app_id")

	if code == "" || state == "" {
		response.Error(c, "Invalid callback", "Missing code or state in callback", http.StatusBadRequest)
		return
	}

	req := &model.OAuthCallbackRequest{
		Provider: provider,
		Code:     code,
		State:    state,
		AppID:    appId,
	}

	loginResp, appID, err := h.userUseCase.HandleOAuthCallback(c.Request.Context(), req)
	if err != nil {
		response.Error(c, "OAuth authentication failed", err.Error(), http.StatusUnauthorized)
		return
	}

	response.Success(c, "OAuth authenticated successfully", gin.H{
		"login_response": loginResp,
		"app_id":         appID,
		"redirect_to":    "/dashboard",
	}, http.StatusOK)
}
