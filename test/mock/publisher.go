package mock

import (
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PublishedMessage struct {
	RoutingKey string
	Body       []byte
}

type MockPublisher struct {
	mu        sync.Mutex
	Messages  []PublishedMessage
	PublishErr error
}

func NewMockPublisher() *MockPublisher {
	return &MockPublisher{}
}

func (m *MockPublisher) Publish(routingKey string, body []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.PublishErr != nil {
		return m.PublishErr
	}

	m.Messages = append(m.Messages, PublishedMessage{
		RoutingKey: routingKey,
		Body:       body,
	})
	return nil
}

func (m *MockPublisher) UpdateConnection(conn *amqp.Connection) error {
	return nil
}

func (m *MockPublisher) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Messages)
}
