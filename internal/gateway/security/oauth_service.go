package security

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"go-gin-clean/internal/entity"
	"go-gin-clean/internal/model"
	"go-gin-clean/pkg/config"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type OAuthService struct {
	googleConfig   *GoogleOAuthConfig
	stateGenerator *StateGenerator
	frontendURLs   map[string]string
	defaultAppID   string
}

type StateGenerator struct {
	secret string
	states map[string]stateInfo
}

type stateInfo struct {
	appID     string
	createdAt time.Time
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

func (sg *StateGenerator) Generate(appID string) string {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.StdEncoding.EncodeToString(b)

	sg.states[state] = stateInfo{
		appID:     appID,
		createdAt: time.Now(),
	}

	return state
}

func (sg *StateGenerator) Validate(state string) (string, bool) {
	info, exists := sg.states[state]
	if !exists {
		return "", false
	}

	if time.Since(info.createdAt) > 10*time.Minute {
		delete(sg.states, state)
		return "", false
	}

	delete(sg.states, state)
	return info.appID, true
}

// CleanupExpiredStates removes expired states
func (sg *StateGenerator) CleanupExpiredStates() {
	for state, info := range sg.states {
		if time.Since(info.createdAt) > 10*time.Minute {
			delete(sg.states, state)
		}
	}
}

// OAuth provider configurations
type GoogleOAuthConfig struct {
	ClientID       string
	ClientSecret   string
	RedirectURL    string
	AuthURL        string
	TokenURL       string
	UserInfoURL    string
	AllowedOrigins []string
}

func NewOAuthService(cfg *config.OAuthConfig) *OAuthService {
	stateGenerator := NewStateGenerator(cfg.OAuthStateString)
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			stateGenerator.CleanupExpiredStates()
		}
	}()

	googleConfig := &GoogleOAuthConfig{
		ClientID:       cfg.GoogleClientID,
		ClientSecret:   cfg.GoogleClientSecret,
		RedirectURL:    cfg.GoogleRedirectURL,
		AuthURL:        "https://accounts.google.com/o/oauth2/auth",
		TokenURL:       "https://oauth2.googleapis.com/token",
		UserInfoURL:    "https://www.googleapis.com/oauth2/v2/userinfo",
		AllowedOrigins: cfg.GoogleAllowedOrigins,
	}

	return &OAuthService{
		googleConfig:   googleConfig,
		stateGenerator: stateGenerator,
		frontendURLs:   cfg.FrontendURLs,
		defaultAppID:   cfg.DefaultAppID,
	}
}

func (s *OAuthService) GetGoogleAuthURL(appID string) string {
	if appID == "" {
		appID = s.defaultAppID
	}

	state := s.stateGenerator.Generate(appID)

	params := url.Values{}
	params.Add("client_id", s.googleConfig.ClientID)
	params.Add("redirect_uri", s.googleConfig.RedirectURL)
	params.Add("response_type", "code")
	params.Add("scope", "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile")
	params.Add("state", state)
	params.Add("access_type", "offline")
	params.Add("prompt", "consent")

	return s.googleConfig.AuthURL + "?" + params.Encode()
}

func (s *OAuthService) HandleGoogleCallback(ctx context.Context, state string, code string) (*entity.User, string, error) {
	appID, valid := s.stateGenerator.Validate(state)
	if !valid {
		return nil, "", errors.New("invalid oauth state")
	}

	tokenResp, err := s.exchangeGoogleCode(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("code exchange failed: %v", err)
	}

	userData, err := s.getGoogleUserData(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user info from Google: %v", err)
	}

	user, err := entity.NewUserFromOAuth(
		userData.Name,
		userData.Email,
		"google",
		userData.ID,
		userData.Picture,
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create user from Google data: %v", err)
	}

	return user, appID, nil
}

func (s *OAuthService) GetFrontendURL(appID string) string {
	if appID == "" {
		appID = s.defaultAppID
	}

	url, exists := s.frontendURLs[appID]
	if !exists {
		url = s.frontendURLs[s.defaultAppID]
	}

	return url
}

func (s *OAuthService) exchangeGoogleCode(ctx context.Context, code string) (*model.TokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", s.googleConfig.ClientID)
	data.Set("client_secret", s.googleConfig.ClientSecret)
	data.Set("redirect_uri", s.googleConfig.RedirectURL)
	data.Set("grant_type", "authorization_code")

	return s.exchangeCode(ctx, s.googleConfig.TokenURL, data)
}

func (s *OAuthService) exchangeCode(ctx context.Context, tokenURL string, data url.Values) (*model.TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp model.TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func (s *OAuthService) getGoogleUserData(ctx context.Context, accessToken string) (*model.GoogleUserData, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.googleConfig.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info with status %d: %s", resp.StatusCode, string(body))
	}

	var userData model.GoogleUserData
	if err := json.Unmarshal(body, &userData); err != nil {
		return nil, err
	}

	return &userData, nil
}

func (s *OAuthService) IsOriginAllowed(provider, origin string) bool {
	var allowedOrigins []string

	switch provider {
	case "google":
		allowedOrigins = s.googleConfig.AllowedOrigins
	default:
		return false
	}

	if len(allowedOrigins) == 0 {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}

	return false
}
