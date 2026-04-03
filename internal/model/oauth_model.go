package model

// User data structures
type GoogleUserData struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// Token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token,omitempty"`
}

// OAuthLoginRequest represents the request to initiate OAuth login
type OAuthLoginRequest struct {
	Provider string `json:"provider" binding:"required,oneof=google facebook microsoft"`
	AppID    string `json:"app_id" binding:"omitempty"`   // Optional: frontend application ID
	Platform string `json:"platform" binding:"omitempty"` // Optional: "web" (default) or "mobile"
}

// OAuthCallbackRequest represents the callback request from OAuth provider
type OAuthCallbackRequest struct {
	Provider string `json:"provider" binding:"required,oneof=google facebook microsoft"`
	Code     string `json:"code" binding:"required"`
	State    string `json:"state" binding:"required"`
	AppID    string `json:"app_id" binding:"omitempty"` // Application ID to redirect back to
}

// OAuthUrlResponse contains the URL to redirect the user to for OAuth login
type OAuthUrlResponse struct {
	AuthURL string `json:"auth_url"`
}
