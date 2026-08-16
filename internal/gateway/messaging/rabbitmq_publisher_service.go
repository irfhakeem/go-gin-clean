package messaging

import (
	"errors"
	"fmt"
	"go-gin-clean/pkg/config"
	"log"
	"sync"

	"github.com/rabbitmq/amqp091-go"
)

const (
	// User events
	UserRegisteredEvent    = "user.register"
	UserResetPasswordEvent = "user.reset_password"
)

type RabbitMQPublisherServiceInterface interface {
	Publish(routingKey string, body []byte) error
	UpdateConnection(conn *amqp091.Connection) error
}

type RabbitMQPublisherService struct {
	conn *amqp091.Connection
	ch   *amqp091.Channel
	cfg  *config.RabbitMQConfig
	mu   sync.Mutex
}

func NewRabbitMQPublisherService(conn *amqp091.Connection, cfg *config.RabbitMQConfig) *RabbitMQPublisherService {
	ch, err := conn.Channel()
	if err != nil {
		return nil
	}
	return &RabbitMQPublisherService{
		conn: conn,
		ch:   ch,
		cfg:  cfg,
	}
}

func (p *RabbitMQPublisherService) Publish(routingKey string, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.publish(routingKey, body)
	if err == nil {
		return nil
	}

	if !isChannelClosed(err) {
		return err
	}

	log.Printf("[RabbitMQPublisherService] channel closed, reconnecting before retry (key=%s): %v", routingKey, err)
	if reconnErr := p.reconnect(); reconnErr != nil {
		return fmt.Errorf("reconnect failed: %w", reconnErr)
	}

	return p.publish(routingKey, body)
}

func (p *RabbitMQPublisherService) UpdateConnection(conn *amqp091.Connection) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conn = conn
	return p.reconnect()
}

func (p *RabbitMQPublisherService) publish(routingKey string, body []byte) error {
	return p.ch.Publish(
		p.cfg.Exchange,
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

func (p *RabbitMQPublisherService) reconnect() error {
	if p.ch != nil {
		p.ch.Close()
	}
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	p.ch = ch
	log.Printf("[RabbitMQPublisherService] channel reconnected successfully")
	return nil
}

func isChannelClosed(err error) bool {
	if errors.Is(err, amqp091.ErrClosed) {
		return true
	}
	var amqpErr *amqp091.Error
	return errors.As(err, &amqpErr) && amqpErr.Code == amqp091.ChannelError
}
