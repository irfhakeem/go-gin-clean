package model

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
	Provider string `json:"provider" binding:"required,oneof=google facebook microsoft"`
	AppID    string `json:"app_id" binding:"omitempty"`
}

type OAuthCallbackRequest struct {
	Provider string `json:"provider" binding:"required,oneof=google facebook microsoft"`
	Code     string `json:"code" binding:"required"`
	State    string `json:"state" binding:"required"`
	AppID    string `json:"app_id" binding:"omitempty"`
}

type OAuthUrlResponse struct {
	AuthURL string `json:"auth_url"`
}

type OAuthClientRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	RedirectURIs string `json:"redirect_uris" binding:"required"`
	Scopes       string `json:"scopes"`
}

type OAuthClientResponse struct {
	PKID         int64  `json:"pkid"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	RedirectURIs string `json:"redirect_uris"`
	Scopes       string `json:"scopes"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
}

type OAuthAuthorizeRequest struct {
	ResponseType        string `form:"response_type" binding:"required,eq=code"`
	ClientID            string `form:"client_id" binding:"required"`
	RedirectURI         string `form:"redirect_uri" binding:"required"`
	Scope               string `form:"scope"`
	State               string `form:"state"`
	CodeChallenge       string `form:"code_challenge"`
	CodeChallengeMethod string `form:"code_challenge_method"`
}

type OAuthTokenRequest struct {
	GrantType    string `form:"grant_type" binding:"required,eq=authorization_code"`
	Code         string `form:"code" binding:"required"`
	RedirectURI  string `form:"redirect_uri" binding:"required"`
	ClientID     string `form:"client_id" binding:"required"`
	ClientSecret string `form:"client_secret" binding:"required"`
	CodeVerifier string `form:"code_verifier"`
}

type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token,omitempty"`
}

type OAuthUserInfoResponse struct {
	Sub      string `json:"sub"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Picture  string `json:"picture"`
	Gender   string `json:"gender"`
	Username string `json:"username"`
	Code     string `json:"code"`
}

type OAuthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
	State            string `json:"state,omitempty"`
}
