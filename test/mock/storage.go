package mock

import (
	"context"
	"fmt"
	"mime/multipart"
	"sync"
)

type MockStorage struct {
	mu           sync.Mutex
	Files        map[string][]byte
	DeletedFiles []string
	UploadErr    error
	DeleteErr    error
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		Files: make(map[string][]byte),
	}
}

func (m *MockStorage) UploadFile(ctx context.Context, filename string, size int64, fileHeader multipart.FileHeader, filePath string) (*string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.UploadErr != nil {
		return nil, m.UploadErr
	}

	url := fmt.Sprintf("https://cdn.example.com/%s%s", filePath, filename)
	m.Files[url] = []byte(filename)
	return &url, nil
}

func (m *MockStorage) DeleteFile(ctx context.Context, path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.DeleteErr != nil {
		return m.DeleteErr
	}

	delete(m.Files, path)
	m.DeletedFiles = append(m.DeletedFiles, path)
	return nil
}
