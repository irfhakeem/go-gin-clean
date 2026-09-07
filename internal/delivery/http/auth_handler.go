package http

import (
	"fmt"
	"net/http"
	"net/url"

	"go-gin-clean/internal/application/usecase"
	"go-gin-clean/internal/delivery/http/response"
	"go-gin-clean/internal/dto"
	pkgerrors "go-gin-clean/pkg/errors"
	"go-gin-clean/pkg/message"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUseCase usecase.AuthUseCase
}

func NewAuthHandler(authUseCase usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.authUseCase.Login(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	setRefreshTokenCookie(c, result.RefreshToken, http.SameSiteStrictMode, 0, false)

	response.Success(c, message.LoginSuccess, gin.H{
		"access_token": result.AccessToken,
	}, http.StatusOK)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.authUseCase.Register(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.RegisterSuccess, nil, http.StatusCreated)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		response.Error(c, pkgerrors.WrapAppError(pkgerrors.Unauthorized, message.ErrTokenNotFound, err))
		return
	}

	result, err := h.authUseCase.RefreshToken(c.Request.Context(), cookie)
	if err != nil {
		response.Error(c, err)
		return
	}

	setRefreshTokenCookie(c, result.RefreshToken, http.SameSiteStrictMode, 0, false)

	response.Success(c, message.RefreshSuccess, result.AccessToken, http.StatusOK)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrUnauthorized))
		return
	}

	if err := h.authUseCase.Logout(c.Request.Context(), userID.(string)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.LogoutSuccess, nil, http.StatusOK)
}

func (h *AuthHandler) SendVerifyEmail(c *gin.Context) {
	var req dto.SendVerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.authUseCase.SendVerifyEmail(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.SendVerificationEmailSuccess, nil, http.StatusOK)
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.authUseCase.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.VerifyEmailSuccess, nil, http.StatusOK)
}

func (h *AuthHandler) SendResetPassword(c *gin.Context) {
	var req dto.SendResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.authUseCase.SendResetPassword(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.SendResetPasswordEmailSuccess, nil, http.StatusOK)
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.authUseCase.ResetPassword(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.ResetPasswordSuccess, nil, http.StatusOK)
}

func (h *AuthHandler) GetOAuthLoginURL(c *gin.Context) {
	var req dto.OAuthLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if req.Platform == "" {
		req.Platform = "web"
	}

	urlResp, err := h.authUseCase.GetOAuthLoginURL(c.Request.Context(), req.Provider, req.AppID, req.Platform)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, message.OAuthLoginSuccess, urlResp, http.StatusOK)
}

func (h *AuthHandler) HandleOAuthCallback(c *gin.Context) {
	provider := c.Param("provider")

	var req dto.OAuthCallbackRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, appID, platform, err := h.authUseCase.HandleOAuthCallback(c.Request.Context(), provider, &req)
	if err != nil {
		reason := response.Reason(c, err)
		if platform == "mobile" {
			deepLink := h.authUseCase.GetOAuthRedirectURL(appID, "mobile")
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
		frontendURL := h.authUseCase.GetOAuthRedirectURL(appID, "web")
		params := url.Values{}
		params.Set("error", reason)
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/oauth/error?%s", frontendURL, params.Encode()))
		return
	}

	if platform == "mobile" {
		deepLink := h.authUseCase.GetOAuthRedirectURL(appID, "mobile")
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

	setRefreshTokenCookie(c, result.RefreshToken, http.SameSiteLaxMode, 7*24*60*60, true)

	frontendURL := h.authUseCase.GetOAuthRedirectURL(appID, "web")
	c.Redirect(http.StatusFound, fmt.Sprintf("%s/oauth/callback#access_token=%s", frontendURL, result.AccessToken))
}

// setRefreshTokenCookie writes the HttpOnly refresh-token cookie.
// Same-site endpoints (login/refresh) use Strict session cookies; the OAuth
// web callback uses a Lax persistent cookie (cross-site redirect from provider).
func setRefreshTokenCookie(c *gin.Context, refreshToken string, sameSite http.SameSite, maxAge int, secure bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		Path:     "/",
		MaxAge:   maxAge,
	})
}
