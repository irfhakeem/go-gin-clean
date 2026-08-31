package http

import (
	"fmt"
	"net/http"
	"net/url"

	"go-gin-clean/internal/application/usecase"
	"go-gin-clean/internal/delivery/http/response"
	"go-gin-clean/internal/dto"
	"go-gin-clean/pkg/message"

	"github.com/gin-gonic/gin"
)

type OAuthHandler struct {
	userUseCase usecase.UserUseCase
}

func NewOAuthHandler(userUseCase usecase.UserUseCase) *OAuthHandler {
	return &OAuthHandler{
		userUseCase: userUseCase,
	}
}

func (h *OAuthHandler) GetLoginURL(c *gin.Context) {
	var req dto.OAuthLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if req.Platform == "" {
		req.Platform = "web"
	}

	urlResp, err := h.userUseCase.GetOAuthLoginURL(c.Request.Context(), req.Provider, req.AppID, req.Platform)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.OAuthLoginSuccess, urlResp, http.StatusOK)
}

func (h *OAuthHandler) CallBack(c *gin.Context) {
	provider := c.Param("provider")

	var query dto.OAuthCallbackRequest
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, appID, platform, err := h.userUseCase.HandleOAuthCallback(c.Request.Context(), provider, &query)
	if err != nil {
		reason := response.Reason(c, err)
		if platform == "mobile" {
			deepLink := h.userUseCase.GetOAuthRedirectURL(appID, "mobile")
			if deepLink == "" {
				response.Error(c, err)
				return
			}

			params := url.Values{}
			params.Set("status", "error")
			params.Set("error", reason)
			c.Redirect(http.StatusFound, fmt.Sprintf("%s?%s", deepLink, params.Encode()))
			return
		}
		frontendURL := h.userUseCase.GetOAuthRedirectURL(appID, "web")
		params := url.Values{}
		params.Set("error", reason)
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/oauth/error?%s", frontendURL, params.Encode()))
		return
	}

	if platform == "mobile" {
		deepLink := h.userUseCase.GetOAuthRedirectURL(appID, "mobile")
		if deepLink == "" {
			return
		}

		params := url.Values{}
		params.Set("status", "success")
		params.Set("access_token", result.AccessToken)
		params.Set("refresh_token", result.RefreshToken)
		c.Redirect(http.StatusFound, fmt.Sprintf("%s?%s", deepLink, params.Encode()))
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
