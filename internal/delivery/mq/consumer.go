package mq

import (
	"context"
	"errors"
	"fmt"
	"net/textproto"
	"time"

	"go-gin-clean/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const (
	MaxRetries = 3
	RetryDelay = 5 * time.Second

	QoS = 1
)

var MainQueues = map[string][]string{
	"email": {"user.register", "user.reset_password"},
}

type EventHandlerFunc func(ctx context.Context, payload []byte) error

type Consumer struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	exchange string
	dlxReady bool

	handlers map[string]EventHandlerFunc
}

func NewConsumer(conn *amqp.Connection, exchange string, handlers map[string]EventHandlerFunc) *Consumer {
	return &Consumer{
		conn:     conn,
		exchange: exchange,
		handlers: handlers,
	}
}

func (c *Consumer) dlxExchange() string            { return c.exchange + ".dlx" }
func (c *Consumer) mainQueueName(q string) string  { return c.exchange + "." + q }
func (c *Consumer) retryQueueName(q string) string { return c.exchange + "." + q + ".retry" }
func (c *Consumer) deadQueueName(q string) string  { return c.exchange + "." + q + ".dead" }
func deadRoutingKey(key string) string             { return key + ".dead" }

func (c *Consumer) Initialize(ctx context.Context) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	c.ch = ch

	if err := c.ch.Qos(QoS, 0, false); err != nil {
		return fmt.Errorf("set qos: %w", err)
	}

	if err := c.ch.ExchangeDeclare(c.exchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}
	if err := c.ch.ExchangeDeclare(c.dlxExchange(), "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlx exchange: %w", err)
	}

	c.dlxReady = true
	for queueName, routingKeys := range MainQueues {
		ok, err := c.declareMainQueue(queueName)
		if err != nil {
			return fmt.Errorf("declare main queue '%s': %w", queueName, err)
		}
		if !ok {
			c.dlxReady = false
			if err := c.ch.QueueBind(c.mainQueueName(queueName), routingKeys[0], c.exchange, false, nil); err != nil {
				return fmt.Errorf("bind degraded queue '%s': %w", queueName, err)
			}
			for _, key := range routingKeys[1:] {
				if err := c.ch.QueueBind(c.mainQueueName(queueName), key, c.exchange, false, nil); err != nil {
					return fmt.Errorf("bind degraded queue '%s' key '%s': %w", queueName, key, err)
				}
			}
			continue
		}

		if err := c.declareRetryQueue(queueName); err != nil {
			return fmt.Errorf("declare retry queue '%s': %w", queueName, err)
		}
		if err := c.declareDeadQueue(queueName); err != nil {
			return fmt.Errorf("declare dead queue '%s': %w", queueName, err)
		}
		if err := c.bindAll(queueName, routingKeys); err != nil {
			return err
		}
	}

	if c.dlxReady {
		logger.Info("mq: retry/DLQ topology declared", zap.String("exchange", c.exchange))
	} else {
		logger.Warn("mq: exchange declared in degraded mode, no retry/DLQ topology", zap.String("exchange", c.exchange))
	}

	return nil
}

func (c *Consumer) declareMainQueue(queueName string) (bool, error) {
	fullName := c.mainQueueName(queueName)

	_, err := c.ch.QueueDeclare(fullName, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange": c.dlxExchange(),
	})
	if err == nil {
		return true, nil
	}

	var amqpErr *amqp.Error
	if !errors.As(err, &amqpErr) || amqpErr.Code != amqp.PreconditionFailed {
		return false, fmt.Errorf("declare queue '%s': %w", fullName, err)
	}

	newCh, chErr := c.conn.Channel()
	if chErr != nil {
		return false, fmt.Errorf("recover channel after 406 on queue '%s': %w", fullName, chErr)
	}
	c.ch = newCh

	if _, err := c.ch.QueueDeclare(fullName, true, false, false, false, nil); err != nil {
		return false, fmt.Errorf("declare queue '%s' (degraded fallback): %w", fullName, err)
	}
	return false, nil
}

func (c *Consumer) declareRetryQueue(queueName string) error {
	_, err := c.ch.QueueDeclare(c.retryQueueName(queueName), true, false, false, false, amqp.Table{
		"x-message-ttl":          int32(RetryDelay.Milliseconds()),
		"x-dead-letter-exchange": c.exchange,
	})
	return err
}

func (c *Consumer) declareDeadQueue(queueName string) error {
	_, err := c.ch.QueueDeclare(c.deadQueueName(queueName), true, false, false, false, nil)
	return err
}

func (c *Consumer) bindAll(queueName string, routingKeys []string) error {
	mainName := c.mainQueueName(queueName)
	retryName := c.retryQueueName(queueName)
	deadName := c.deadQueueName(queueName)

	for _, key := range routingKeys {
		if err := c.ch.QueueBind(mainName, key, c.exchange, false, nil); err != nil {
			return fmt.Errorf("bind main queue '%s' key '%s': %w", mainName, key, err)
		}
		if err := c.ch.QueueBind(retryName, key, c.dlxExchange(), false, nil); err != nil {
			return fmt.Errorf("bind retry queue '%s' key '%s': %w", retryName, key, err)
		}
		if err := c.ch.QueueBind(deadName, deadRoutingKey(key), c.dlxExchange(), false, nil); err != nil {
			return fmt.Errorf("bind dead queue '%s' key '%s': %w", deadName, key, err)
		}
	}
	return nil
}

func (c *Consumer) Consume(ctx context.Context, queueName string) error {
	deliveries, err := c.ch.Consume(c.mainQueueName(queueName), "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume queue '%s': %w", queueName, err)
	}

	retryName := c.retryQueueName(queueName)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-deliveries:
				if !ok {
					return
				}
				c.handleDelivery(ctx, d, retryName)
			}
		}
	}()

	return nil
}

func (c *Consumer) handleDelivery(ctx context.Context, d amqp.Delivery, retryQueue string) {
	handler, ok := c.handlers[d.RoutingKey]
	if !ok {
		logger.Warn("mq: no handler registered for routing key, dropping", zap.String("routing_key", d.RoutingKey))
		_ = d.Nack(false, false)
		return
	}

	err := handler(ctx, d.Body)
	if err == nil {
		_ = d.Ack(false)
		return
	}

	if IsRetryable(err) && retryCount(d, retryQueue) < MaxRetries {
		logger.Warn("mq: retryable error, sending to retry queue", zap.String("routing_key", d.RoutingKey), zap.Error(err))
		_ = d.Nack(false, false)
		return
	}

	logger.Error("mq: permanent failure, parking in dead queue", zap.String("routing_key", d.RoutingKey), zap.Error(err))
	if pubErr := c.publishDead(d); pubErr != nil {
		logger.Error("mq: failed to park dead message, requeueing as last resort", zap.Error(pubErr))
		_ = d.Nack(false, true)
		return
	}
	_ = d.Ack(false)
}

func (c *Consumer) publishDead(d amqp.Delivery) error {
	return c.ch.Publish(c.dlxExchange(), deadRoutingKey(d.RoutingKey), false, false, amqp.Publishing{
		ContentType:  d.ContentType,
		DeliveryMode: amqp.Persistent,
		Headers:      d.Headers,
		Body:         d.Body,
	})
}

func retryCount(d amqp.Delivery, retryQueue string) int {
	raw, ok := d.Headers["x-death"]
	if !ok {
		return 0
	}
	deaths, ok := raw.([]any)
	if !ok {
		return 0
	}
	for _, entry := range deaths {
		table, ok := entry.(amqp.Table)
		if !ok {
			continue
		}
		if queue, _ := table["queue"].(string); queue != retryQueue {
			continue
		}
		if count, ok := table["count"].(int64); ok {
			return int(count)
		}
	}
	return 0
}

type nonRetryableError struct{ error }

func (e *nonRetryableError) Unwrap() error { return e.error }

func NonRetryable(err error) error {
	if err == nil {
		return nil
	}
	return &nonRetryableError{err}
}

func IsRetryable(err error) bool {
	var nre *nonRetryableError
	if errors.As(err, &nre) {
		return false
	}

	var smtpErr *textproto.Error
	if errors.As(err, &smtpErr) {
		return smtpErr.Code < 500
	}

	return true
}
