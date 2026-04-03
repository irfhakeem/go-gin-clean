package http

import (
	"fmt"
	"go-gin-clean/internal/delivery/http/response"
	"go-gin-clean/internal/model"
	"go-gin-clean/internal/model/validator"
	"go-gin-clean/internal/usecase"
	"net/http"
	"net/url"

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
		response.Error(c, "Failed to parse JSON", validator.BuildValidationMessage(err), http.StatusBadRequest)
		return
	}

	if req.Platform == "" {
		req.Platform = "web"
	}

	urlResp, err := h.userUseCase.GetOAuthLoginURL(c.Request.Context(), req.Provider, req.AppID, req.Platform)
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

	if code == "" || state == "" {
		response.Error(c, "Invalid callback", "Missing code or state in callback", http.StatusBadRequest)
		return
	}

	req := &model.OAuthCallbackRequest{
		Provider: provider,
		Code:     code,
		State:    state,
	}

	result, appID, platform, err := h.userUseCase.HandleOAuthCallback(c.Request.Context(), req)
	if err != nil {
		if platform == "mobile" {
			h.redirectMobileError(c, appID, err.Error())
			return
		}
		h.redirectWebError(c, appID, err.Error())
		return
	}

	if platform == "mobile" {
		h.redirectMobileSuccess(c, appID, result.AccessToken, result.RefreshToken)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    result.RefreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60,
	})

	frontendURL := h.userUseCase.GetOAuthRedirectURL(appID, "web")
	c.Redirect(http.StatusFound, fmt.Sprintf("%s/oauth/callback#access_token=%s", frontendURL, result.AccessToken))
}

// Helper
func (h *OAuthHandler) redirectWebError(c *gin.Context, appID, reason string) {
	frontendURL := h.userUseCase.GetOAuthRedirectURL(appID, "web")
	params := url.Values{}
	params.Set("error", reason)
	c.Redirect(http.StatusFound, fmt.Sprintf("%s/oauth/error?%s", frontendURL, params.Encode()))
}

func (h *OAuthHandler) redirectMobileSuccess(c *gin.Context, appID, accessToken, refreshToken string) {
	deepLink := h.userUseCase.GetOAuthRedirectURL(appID, "mobile")
	if deepLink == "" {
		response.Error(c, "Mobile OAuth not configured", "deep link URI is not configured for this app", http.StatusInternalServerError)
		return
	}

	params := url.Values{}
	params.Set("status", "success")
	params.Set("access_token", accessToken)
	params.Set("refresh_token", refreshToken)
	c.Redirect(http.StatusFound, fmt.Sprintf("%s?%s", deepLink, params.Encode()))
}

func (h *OAuthHandler) redirectMobileError(c *gin.Context, appID, reason string) {
	deepLink := h.userUseCase.GetOAuthRedirectURL(appID, "mobile")
	if deepLink == "" {
		response.Error(c, "OAuth authentication failed", reason, http.StatusUnauthorized)
		return
	}

	params := url.Values{}
	params.Set("status", "error")
	params.Set("error", reason)
	c.Redirect(http.StatusFound, fmt.Sprintf("%s?%s", deepLink, params.Encode()))
}
