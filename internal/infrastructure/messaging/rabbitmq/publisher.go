package rabbitmq

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"go-gin-clean/internal/application/port"
	"go-gin-clean/pkg/config"

	amqp091 "github.com/rabbitmq/amqp091-go"
)

type publisher struct {
	conn *amqp091.Connection
	ch   *amqp091.Channel
	cfg  *config.RabbitMQConfig
	mu   sync.Mutex
}

func NewPublisher(conn *amqp091.Connection, cfg *config.RabbitMQConfig) port.Publisher {
	ch, err := conn.Channel()
	if err != nil {
		return nil
	}
	return &publisher{
		conn: conn,
		ch:   ch,
		cfg:  cfg,
	}
}

var _ port.Publisher = (*publisher)(nil)

func (p *publisher) Publish(routingKey string, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.publish(routingKey, body)
	if err == nil {
		return nil
	}

	if errors.Is(err, amqp091.ErrClosed) {
		return fmt.Errorf("channel closed: %w", err)
	}

	var amqpErr *amqp091.Error
	if errors.As(err, &amqpErr) && amqpErr.Code == amqp091.ChannelError {
		return fmt.Errorf("channel error: %w", err)
	}

	log.Printf("[rabbitMQPublisher] channel closed, reconnecting before retry (key=%s): %v", routingKey, err)
	if reconnErr := p.reconnect(); reconnErr != nil {
		return fmt.Errorf("reconnect failed: %w", reconnErr)
	}

	return p.publish(routingKey, body)
}

func (p *publisher) UpdateConnection(conn *amqp091.Connection) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conn = conn
	return p.reconnect()
}

func (p *publisher) publish(routingKey string, body []byte) error {
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

func (p *publisher) reconnect() error {
	if p.ch != nil {
		p.ch.Close()
	}
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	p.ch = ch
	log.Printf("[publisher] channel reconnected successfully")
	return nil
}
