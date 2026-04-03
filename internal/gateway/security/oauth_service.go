package security

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

	"go-gin-clean/internal/entity"
	"go-gin-clean/internal/model"
	"go-gin-clean/pkg/config"
)

type OAuthServiceInterface interface {
	GetGoogleAuthURL(appID string, platform string) string
	HandleGoogleCallback(ctx context.Context, state string, code string) (*entity.User, string, string, error)
	GetFrontendURL(appID string) string
	GetMobileDeepLinkURL(appID string) string
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type OAuthService struct {
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
	WebRedirectURL string // backend registered at gcc auth
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
	redirectURI string // the exact redirect_uri sent in the authorisation request
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

// Generate creates a new opaque state token and stores its metadata.
// The redirectURI is stored server-side so the token-exchange step uses the
// exact same URI that was in the authorisation request.
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

// Validate checks the state token and returns the stored metadata.
// The token is consumed immediately (single-use).
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

// CleanupExpiredStates removes stale tokens. Called by a background goroutine.
func (sg *StateGenerator) CleanupExpiredStates() {
	sg.mu.Lock()
	defer sg.mu.Unlock()

	for state, info := range sg.states {
		if time.Since(info.createdAt) > 10*time.Minute {
			delete(sg.states, state)
		}
	}
}

func NewOAuthService(cfg *config.OAuthConfig, httpClients ...HTTPClient) *OAuthService {
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

	return &OAuthService{
		googleConfig:    googleCfg,
		stateGenerator:  sg,
		frontendURLs:    cfg.FrontendURLs,
		mobileDeepLinks: cfg.MobileDeepLinks,
		defaultAppID:    cfg.DefaultAppID,
		httpClient:      hc,
	}
}

func (s *OAuthService) GetGoogleAuthURL(appID, platform string) string {
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

func (s *OAuthService) HandleGoogleCallback(ctx context.Context, state, code string) (*entity.User, string, string, error) {
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

func (s *OAuthService) GetFrontendURL(appID string) string {
	if appID == "" {
		appID = s.defaultAppID
	}
	if u, ok := s.frontendURLs[appID]; ok {
		return u
	}
	return s.frontendURLs[s.defaultAppID]
}

func (s *OAuthService) GetMobileDeepLinkURL(appID string) string {
	if appID == "" {
		appID = s.defaultAppID
	}
	if u, ok := s.mobileDeepLinks[appID]; ok {
		return u
	}
	return s.mobileDeepLinks[s.defaultAppID]
}

func (s *OAuthService) IsOriginAllowed(provider, origin string) bool {
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
func (s *OAuthService) buildTokenPayload(code, redirectURI string) url.Values {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", s.googleConfig.ClientID)
	data.Set("client_secret", s.googleConfig.ClientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")
	return data
}

func (s *OAuthService) exchangeTokens(ctx context.Context, payload url.Values) (*model.TokenResponse, error) {
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

	var tokenResp model.TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}
	return &tokenResp, nil
}

func (s *OAuthService) googleUserToEntity(ctx context.Context, accessToken string) (*entity.User, error) {
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

func (s *OAuthService) getGoogleUserData(ctx context.Context, accessToken string) (*model.GoogleUserData, error) {
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

	var userData model.GoogleUserData
	if err := json.Unmarshal(body, &userData); err != nil {
		return nil, err
	}
	return &userData, nil
}
