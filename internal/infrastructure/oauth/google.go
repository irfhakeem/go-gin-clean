package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go-gin-clean/internal/application/port"
	"go-gin-clean/internal/domain/entity"
	"go-gin-clean/internal/dto"

	"go-gin-clean/pkg/config"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type oauthService struct {
	googleConfig    *googleOAuthConfig
	stateGenerator  *StateGenerator
	frontendURLs    map[string]string
	mobileDeepLinks map[string]string
	defaultAppID    string
	httpClient      HTTPClient
}

type googleOAuthConfig struct {
	ClientID       string
	ClientSecret   string
	WebRedirectURL string
	AuthURL        string
	TokenURL       string
	UserInfoURL    string
	AllowedOrigins []string
}

type StateGenerator struct {
	secret string
	states map[string]stateInfo
	mu     sync.Mutex
}

type stateInfo struct {
	appID       string
	platform    string
	redirectURI string
	createdAt   time.Time
}

func NewStateGenerator(secret string) *StateGenerator {
	if secret == "" {
		b := make([]byte, 32)
		rand.Read(b)
		secret = base64.StdEncoding.EncodeToString(b)
	}
	return &StateGenerator{
		secret: secret,
		states: make(map[string]stateInfo),
	}
}

func (sg *StateGenerator) Generate(appID, platform, redirectURI string) string {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.StdEncoding.EncodeToString(b)

	if platform == "" {
		platform = "web"
	}

	sg.mu.Lock()
	defer sg.mu.Unlock()

	sg.states[state] = stateInfo{
		appID:       appID,
		platform:    platform,
		redirectURI: redirectURI,
		createdAt:   time.Now(),
	}

	return state
}

func (sg *StateGenerator) Validate(state string) (appID, platform, redirectURI string, valid bool) {
	sg.mu.Lock()
	defer sg.mu.Unlock()

	info, exists := sg.states[state]
	if !exists {
		return "", "", "", false
	}

	if time.Since(info.createdAt) > 10*time.Minute {
		delete(sg.states, state)
		return "", "", "", false
	}

	delete(sg.states, state)
	return info.appID, info.platform, info.redirectURI, true
}

func (sg *StateGenerator) CleanupExpiredStates() {
	sg.mu.Lock()
	defer sg.mu.Unlock()

	for state, info := range sg.states {
		if time.Since(info.createdAt) > 10*time.Minute {
			delete(sg.states, state)
		}
	}
}

func NewGoogleOAuth(cfg *config.OAuthConfig, httpClients ...HTTPClient) port.OAuthProvider {
	sg := NewStateGenerator(cfg.OAuthStateString)
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			sg.CleanupExpiredStates()
		}
	}()

	googleCfg := &googleOAuthConfig{
		ClientID:       cfg.GoogleClientID,
		ClientSecret:   cfg.GoogleClientSecret,
		WebRedirectURL: cfg.GoogleRedirectURL,
		AuthURL:        "https://accounts.google.com/o/oauth2/auth",
		TokenURL:       "https://oauth2.googleapis.com/token",
		UserInfoURL:    "https://www.googleapis.com/oauth2/v2/userinfo",
		AllowedOrigins: cfg.GoogleAllowedOrigins,
	}

	hc := HTTPClient(&http.Client{})
	if len(httpClients) > 0 {
		hc = httpClients[0]
	}

	return &oauthService{
		googleConfig:    googleCfg,
		stateGenerator:  sg,
		frontendURLs:    cfg.FrontendURLs,
		mobileDeepLinks: cfg.MobileDeepLinks,
		defaultAppID:    cfg.DefaultAppID,
		httpClient:      hc,
	}
}

func (s *oauthService) GetGoogleAuthURL(appID, platform string) string {
	if appID == "" {
		appID = s.defaultAppID
	}

	redirectURI := s.googleConfig.WebRedirectURL
	state := s.stateGenerator.Generate(appID, platform, redirectURI)

	params := url.Values{}
	params.Add("client_id", s.googleConfig.ClientID)
	params.Add("redirect_uri", redirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile")
	params.Add("state", state)
	params.Add("access_type", "offline")
	params.Add("prompt", "consent")

	return s.googleConfig.AuthURL + "?" + params.Encode()
}

func (s *oauthService) HandleGoogleCallback(ctx context.Context, state, code string) (*entity.User, string, string, error) {
	appID, platform, redirectURI, valid := s.stateGenerator.Validate(state)
	if !valid {
		return nil, "", "", errors.New("invalid oauth state")
	}

	tokenResp, err := s.exchangeTokens(ctx, s.buildTokenPayload(code, redirectURI))
	if err != nil {
		return nil, "", "", fmt.Errorf("code exchange failed: %w", err)
	}

	user, err := s.googleUserToEntity(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, "", "", err
	}

	return user, appID, platform, nil
}

func (s *oauthService) GetFrontendURL(appID string) string {
	if appID == "" {
		appID = s.defaultAppID
	}
	if u, ok := s.frontendURLs[appID]; ok {
		return u
	}
	return s.frontendURLs[s.defaultAppID]
}

func (s *oauthService) GetMobileDeepLinkURL(appID string) string {
	if appID == "" {
		appID = s.defaultAppID
	}
	if u, ok := s.mobileDeepLinks[appID]; ok {
		return u
	}
	return s.mobileDeepLinks[s.defaultAppID]
}

func (s *oauthService) IsOriginAllowed(provider, origin string) bool {
	var allowed []string
	switch provider {
	case "google":
		allowed = s.googleConfig.AllowedOrigins
	default:
		return false
	}
	for _, a := range allowed {
		if a == "*" || a == origin {
			return true
		}
	}
	return false
}

// Helper
func (s *oauthService) buildTokenPayload(code, redirectURI string) url.Values {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", s.googleConfig.ClientID)
	data.Set("client_secret", s.googleConfig.ClientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")
	return data
}

func (s *oauthService) exchangeTokens(ctx context.Context, payload url.Values) (*dto.TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.googleConfig.TokenURL, strings.NewReader(payload.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, body)
	}

	var tokenResp dto.TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}
	return &tokenResp, nil
}

func (s *oauthService) googleUserToEntity(ctx context.Context, accessToken string) (*entity.User, error) {
	userData, err := s.getGoogleUserData(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info from Google: %w", err)
	}
	user, err := entity.NewUserFromOAuth(userData.Name, userData.Email, "google", userData.ID, userData.Picture)
	if err != nil {
		return nil, fmt.Errorf("failed to build user from Google data: %w", err)
	}
	return user, nil
}

func (s *oauthService) getGoogleUserData(ctx context.Context, accessToken string) (*dto.GoogleUserData, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.googleConfig.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info with status %d: %s", resp.StatusCode, body)
	}

	var userData dto.GoogleUserData
	if err := json.Unmarshal(body, &userData); err != nil {
		return nil, err
	}
	return &userData, nil
}
