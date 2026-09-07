package dto

type (
	LoginRequest struct {
		Email    string `json:"email"    binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	LoginResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	RegisterRequest struct {
		Name     string `json:"name"     binding:"required,min=2,max=100"`
		Email    string `json:"email"    binding:"required,email,max=254"`
		Password string `json:"password" binding:"required,password"`
	}

	RefreshTokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	VerifyEmailRequest struct {
		Token string `json:"token" binding:"required"`
	}

	SendVerifyEmailRequest struct {
		Email string `json:"email" binding:"required,email"`
	}

	SendResetPasswordRequest struct {
		Email string `json:"email" binding:"required,email"`
	}

	ResetPasswordRequest struct {
		Token       string `json:"token"        binding:"required"`
		NewPassword string `json:"new_password" binding:"required,password"`
	}

	OAuthLoginRequest struct {
		Provider string `json:"provider" binding:"required,oneof=google"`
		AppID    string `json:"app_id"   binding:"omitempty"`
		Platform string `json:"platform" binding:"omitempty,oneof=web mobile"`
	}

	OAuthCallbackRequest struct {
		Code  string `json:"code"  binding:"required"`
		State string `json:"state" binding:"required"`
	}

	OAuthUrlResponse struct {
		AuthURL string `json:"auth_url"`
	}

	GoogleUserData struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
	}

	TokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		IDToken      string `json:"id_token,omitempty"`
	}
)

func FormatLoginResponse(accessToken, refreshToken string) *LoginResponse {
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}
