package mock

import (
	"fmt"
	"sync"
)

type SentEmail struct {
	To      string
	Subject string
	Body    string
}

type MockMailer struct {
	mu      sync.Mutex
	Sent    []SentEmail
	SendErr error
}

func NewMockMailer() *MockMailer {
	return &MockMailer{}
}

func (m *MockMailer) SendEmail(to, subject, body string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.SendErr != nil {
		return m.SendErr
	}

	m.Sent = append(m.Sent, SentEmail{
		To:      to,
		Subject: subject,
		Body:    body,
	})
	return nil
}

func (m *MockMailer) LoadTemplate(templateName string, data any) (string, error) {
	return fmt.Sprintf("<html><body>Template: %s, Data: %v</body></html>", templateName, data), nil
}

func (m *MockMailer) LastSent() *SentEmail {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Sent) == 0 {
		return nil
	}
	return &m.Sent[len(m.Sent)-1]
}
