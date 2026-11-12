package mock

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	pkgerrors "go-gin-clean/pkg/errors"
)

type MockCache struct {
	mu    sync.Mutex
	store map[string]cacheItem
}

type cacheItem struct {
	value     string
	expiresAt time.Time
}

func NewMockCache() *MockCache {
	return &MockCache{
		store: make(map[string]cacheItem),
	}
}
func (m *MockCache) SetWithExpiration(ctx context.Context, key string, value any, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var str string
	switch v := value.(type) {
	case string:
		str = v
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return err
		}
		str = string(bytes)
	}

	var exp time.Time
	if expiration > 0 {
		exp = time.Now().Add(expiration)
	}

	m.store[key] = cacheItem{value: str, expiresAt: exp}
	return nil
}

func (m *MockCache) Get(ctx context.Context, key string, dest any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.store[key]
	if !exists {
		return pkgerrors.ErrCacheMiss
	}

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		delete(m.store, key)
		return pkgerrors.ErrCacheMiss
	}

	return json.Unmarshal([]byte(item.value), dest)
}
func (m *MockCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.store, key)
	return nil
}

func (m *MockCache) DeletePattern(ctx context.Context, pattern string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	prefix := strings.TrimSuffix(pattern, "*")
	for k := range m.store {
		if strings.HasPrefix(k, prefix) {
			delete(m.store, k)
		}
	}
	return nil
}
func (m *MockCache) Flush() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store = make(map[string]cacheItem)
}
