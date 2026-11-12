package mock

import (
	"context"
	"errors"
	"sync"
	"time"

	"go-gin-clean/internal/application/port/outbound"
	"go-gin-clean/internal/domain/entity"

	"github.com/google/uuid"
)

type MockTransactionManager struct {
	mu         sync.Mutex
	Called     bool
	RolledBack bool
	ShouldErr  error
}

func NewMockTransactionManager() *MockTransactionManager {
	return &MockTransactionManager{}
}

func (m *MockTransactionManager) WithTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	m.mu.Lock()
	m.Called = true
	m.mu.Unlock()

	if m.ShouldErr != nil {
		m.mu.Lock()
		m.RolledBack = true
		m.mu.Unlock()
		return m.ShouldErr
	}

	err := fn(ctx)
	if err != nil {
		m.mu.Lock()
		m.RolledBack = true
		m.mu.Unlock()
		return err
	}
	return nil
}

type MockUserRepository struct {
	mu          sync.Mutex
	Users       map[string]*entity.User
	CreateErr   error
	FindErr     error
	UpdateErr   error
	DeleteErr   error
	ExistResult bool
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		Users: make(map[string]*entity.User),
	}
}

func (m *MockUserRepository) FindAll(ctx context.Context, params outbound.FindAllUsersParams) ([]*entity.User, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.FindErr != nil {
		return nil, 0, m.FindErr
	}
	result := make([]*entity.User, 0, len(m.Users))
	for _, u := range m.Users {
		result = append(result, u)
	}
	return result, int64(len(result)), nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.FindErr != nil {
		return nil, m.FindErr
	}
	u, ok := m.Users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.FindErr != nil {
		return nil, m.FindErr
	}
	for _, u := range m.Users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *MockUserRepository) ExistByEmail(ctx context.Context, email string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.Users {
		if u.Email == email {
			return true
		}
	}
	return m.ExistResult
}

func (m *MockUserRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.CreateErr != nil {
		return nil, m.CreateErr
	}
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m.Users[user.ID.String()] = user
	return user, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.UpdateErr != nil {
		return nil, m.UpdateErr
	}
	m.Users[user.ID.String()] = user
	return user, nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	delete(m.Users, id)
	return nil
}

type MockRefreshTokenRepository struct {
	mu            sync.Mutex
	Tokens        map[string]*entity.RefreshToken
	RevokedUsers  []string
	RevokedTokens []string
}

func NewMockRefreshTokenRepository() *MockRefreshTokenRepository {
	return &MockRefreshTokenRepository{
		Tokens: make(map[string]*entity.RefreshToken),
	}
}

func (m *MockRefreshTokenRepository) Save(ctx context.Context, token *entity.RefreshToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Tokens[token.Token] = token
	return nil
}
func (m *MockRefreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RevokedUsers = append(m.RevokedUsers, userID)
	for _, t := range m.Tokens {
		if t.UserID.String() == userID {
			t.IsRevoked = true
		}
	}
	return nil
}

func (m *MockRefreshTokenRepository) RevokeByToken(ctx context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RevokedTokens = append(m.RevokedTokens, token)
	if t, ok := m.Tokens[token]; ok {
		t.IsRevoked = true
	}
	return nil
}
func (m *MockRefreshTokenRepository) IsTokenValid(ctx context.Context, token string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.Tokens[token]
	if !ok {
		return false
	}
	return !t.IsRevoked && t.ExpiryAt.After(time.Now())
}

type MockOutboxRepository struct {
	mu           sync.Mutex
	Messages     []*entity.Outbox
	CreateErr    error
	PublishedIDs []int64
}

func NewMockOutboxRepository() *MockOutboxRepository {
	return &MockOutboxRepository{}
}

func (m *MockOutboxRepository) Create(ctx context.Context, msg *entity.Outbox) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.CreateErr != nil {
		return m.CreateErr
	}
	msg.ID = int64(len(m.Messages) + 1)
	m.Messages = append(m.Messages, msg)
	return nil
}

func (m *MockOutboxRepository) FetchAndLockPending(ctx context.Context, limit int) ([]*entity.Outbox, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var pending []*entity.Outbox
	for _, msg := range m.Messages {
		if msg.Status == entity.OutboxStatusPending {
			pending = append(pending, msg)
			if len(pending) >= limit {
				break
			}
		}
	}
	return pending, nil
}

func (m *MockOutboxRepository) MarkPublished(ctx context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PublishedIDs = append(m.PublishedIDs, id)
	for _, msg := range m.Messages {
		if msg.ID == id {
			msg.Status = entity.OutboxStatusPublished
		}
	}
	return nil
}

func (m *MockOutboxRepository) RetryOrFail(ctx context.Context, id int64, errMsg string) error {
	return nil
}

func (m *MockOutboxRepository) ResetStuck(ctx context.Context, stuckDuration time.Duration) error {
	return nil
}
