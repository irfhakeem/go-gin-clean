package http

import (
	"fmt"
	"net/http"
	"net/url"

	"go-gin-clean/internal/model"
	"go-gin-clean/internal/usecase"
	pkgerror "go-gin-clean/pkg/error"
	"go-gin-clean/pkg/response"

	"github.com/gin-gonic/gin"
)

type OAuthHandler struct {
	userUseCase usecase.UserUseCaseInterface
}

func NewOAuthHandler(userUseCase usecase.UserUseCaseInterface) *OAuthHandler {
	return &OAuthHandler{
		userUseCase: userUseCase,
	}
}

func (h *OAuthHandler) GetLoginURL(c *gin.Context) {
	var req model.OAuthLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		bindError(c, err)
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

	response.Success(c, pkgerror.OAuthLoginSuccess, urlResp, http.StatusOK)
}

func (h *OAuthHandler) CallBack(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		response.Error(c, pkgerror.BadRequest(pkgerror.ErrOAuthCallback, nil))
		return
	}

	req := &model.OAuthCallbackRequest{
		Provider: provider,
		Code:     code,
		State:    state,
	}

	result, appID, platform, err := h.userUseCase.HandleOAuthCallback(c.Request.Context(), req)
	if err != nil {
		reason := response.Reason(c, err)
		if platform == "mobile" {
			deepLink := h.userUseCase.GetOAuthRedirectURL(appID, "mobile")
			if deepLink == "" {
				response.Error(c, pkgerror.Unauthorized(pkgerror.ErrOAuthCallback, fmt.Errorf("%s", reason)))
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
			response.Error(c, pkgerror.InternalServerError(pkgerror.ErrOAuthCallback, nil))
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
