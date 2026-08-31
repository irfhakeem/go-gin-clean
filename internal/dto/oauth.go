package dto

type GoogleUserData struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token,omitempty"`
}

type OAuthLoginRequest struct {
	Provider string `json:"provider" binding:"required,oneof=google"`
	AppID    string `json:"app_id"   binding:"omitempty"`
	Platform string `json:"platform" binding:"omitempty,oneof=web mobile"`
}

type OAuthCallbackRequest struct {
	Code  string `json:"code"     binding:"required"`
	State string `json:"state"    binding:"required"`
	AppID string `json:"app_id"   binding:"omitempty"`
}

type OAuthUrlResponse struct {
	AuthURL string `json:"auth_url"`
}
