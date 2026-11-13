package mock

import (
	"errors"
	"strings"
	"sync"
	"time"

	"go-gin-clean/internal/domain/auth"

	"github.com/google/uuid"
)

type MockTokenMaker struct {
	mu           sync.Mutex
	AccessTokens map[string]uuid.UUID
	GenerateErr  error
	ValidateErr  error
}

func NewMockTokenMaker() *MockTokenMaker {
	return &MockTokenMaker{
		AccessTokens: make(map[string]uuid.UUID),
	}
}

func (m *MockTokenMaker) GenerateAccessToken(id uuid.UUID, role string) (string, time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.GenerateErr != nil {
		return "", time.Time{}, m.GenerateErr
	}

	token := "access_token_" + id.String()
	m.AccessTokens[token] = id
	return token, time.Now().Add(1 * time.Hour), nil
}

func (m *MockTokenMaker) GenerateRefreshToken(id uuid.UUID) (string, time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.GenerateErr != nil {
		return "", time.Time{}, m.GenerateErr
	}

	token := "refresh_token_" + id.String()
	return token, time.Now().Add(7 * 24 * time.Hour), nil
}

func (m *MockTokenMaker) ValidateAccessToken(tokenStr string) (*auth.AccessTokenClaims, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ValidateErr != nil {
		return nil, m.ValidateErr
	}

	id, ok := m.AccessTokens[tokenStr]
	if !ok {
		// Fallback for mock tokens
		if strings.HasPrefix(tokenStr, "access_token_") {
			parsed, err := uuid.Parse(strings.TrimPrefix(tokenStr, "access_token_"))
			if err == nil {
				id = parsed
				ok = true
			}
		}
	}

	if !ok {
		return nil, errors.New("invalid access token")
	}

	return &auth.AccessTokenClaims{
		UserID:    id,
		UserRole:  "user",
		TokenType: "access",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}, nil
}

func (m *MockTokenMaker) ValidateRefreshToken(tokenStr string) (*auth.RefreshTokenClaims, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ValidateErr != nil {
		return nil, m.ValidateErr
	}

	if !strings.HasPrefix(tokenStr, "refresh_token_") {
		return nil, errors.New("invalid refresh token")
	}

	id, err := uuid.Parse(strings.TrimPrefix(tokenStr, "refresh_token_"))
	if err != nil {
		return nil, err
	}

	return &auth.RefreshTokenClaims{
		UserID:    id,
		TokenType: "refresh",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}, nil
}

type MockHasher struct {
	HashErr     error
	ValidateErr error
}

func NewMockHasher() *MockHasher {
	return &MockHasher{}
}

func (m *MockHasher) HashPassword(password string) (string, error) {
	if m.HashErr != nil {
		return "", m.HashErr
	}
	return "mock_hash_" + password, nil
}

func (m *MockHasher) ValidatePassword(password, hash string) error {
	if m.ValidateErr != nil {
		return m.ValidateErr
	}
	if hash != "mock_hash_"+password && !strings.HasSuffix(hash, password) {
		return errors.New("password does not match")
	}
	return nil
}

type MockEncryptor struct {
	EncryptErr error
	DecryptErr error
}

func NewMockEncryptor() *MockEncryptor {
	return &MockEncryptor{}
}

func (m *MockEncryptor) EncryptInternal(plaintext string) (string, error) {
	if m.EncryptErr != nil {
		return "", m.EncryptErr
	}
	return "enc_" + plaintext, nil
}

func (m *MockEncryptor) DecryptInternal(ciphertext string) (string, error) {
	if m.DecryptErr != nil {
		return "", m.DecryptErr
	}
	return strings.TrimPrefix(ciphertext, "enc_"), nil
}

func (m *MockEncryptor) EncryptURLSafe(plaintext string) (string, error) {
	if m.EncryptErr != nil {
		return "", m.EncryptErr
	}
	return "urlsafe_" + plaintext, nil
}

func (m *MockEncryptor) DecryptURLSafe(ciphertext string) (string, error) {
	if m.DecryptErr != nil {
		return "", m.DecryptErr
	}
	return strings.TrimPrefix(ciphertext, "urlsafe_"), nil
}
