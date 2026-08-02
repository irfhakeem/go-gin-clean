package messaging

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/rabbitmq/amqp091-go"
)

const (
	// User events
	UserRegisteredEvent    = "user.register"
	UserResetPasswordEvent = "user.reset_password"
)

type PublisherServiceInterface interface {
	Publish(routingKey string, body []byte) error
	UpdateConnection(conn *amqp091.Connection) error
}

type PublisherService struct {
	conn     *amqp091.Connection
	ch       *amqp091.Channel
	exchange string
	mu       sync.Mutex
}

func NewPublisherService(conn *amqp091.Connection, exchange string) *PublisherService {
	ch, err := conn.Channel()
	if err != nil {
		return nil
	}
	return &PublisherService{
		conn:     conn,
		ch:       ch,
		exchange: exchange,
	}
}

func (p *PublisherService) Publish(routingKey string, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.publish(routingKey, body)
	if err == nil {
		return nil
	}

	if !isChannelClosed(err) {
		return err
	}

	log.Printf("[PublisherService] channel closed, reconnecting before retry (key=%s): %v", routingKey, err)
	if reconnErr := p.reconnect(); reconnErr != nil {
		return fmt.Errorf("reconnect failed: %w", reconnErr)
	}

	return p.publish(routingKey, body)
}

func (p *PublisherService) UpdateConnection(conn *amqp091.Connection) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conn = conn
	return p.reconnect()
}

func (p *PublisherService) publish(routingKey string, body []byte) error {
	return p.ch.Publish(
		p.exchange,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         body,
		},
	)
}

func (p *PublisherService) reconnect() error {
	if p.ch != nil {
		p.ch.Close()
	}
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	p.ch = ch
	log.Printf("[PublisherService] channel reconnected successfully")
	return nil
}

func isChannelClosed(err error) bool {
	if errors.Is(err, amqp091.ErrClosed) {
		return true
	}
	var amqpErr *amqp091.Error
	return errors.As(err, &amqpErr) && amqpErr.Code == amqp091.ChannelError
}
