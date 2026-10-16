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
)

func FormatLoginResponse(accessToken, refreshToken string) *LoginResponse {
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}
