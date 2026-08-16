package connection

import (
	"context"
	"time"

	"go-gin-clean/pkg/config"
	"go-gin-clean/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

func ConnectRabbitMQ(cfg *config.RabbitMQConfig) (*amqp.Connection, error) {
	return amqp.Dial(cfg.Url)
}

func CloseRabbitMQ(conn *amqp.Connection) error {
	if conn == nil {
		return nil
	}
	return conn.Close()
}

func WatchRabbitMQ(ctx context.Context, conn *amqp.Connection, cfg *config.RabbitMQConfig, onReconnect func(newConn *amqp.Connection)) {
	const (
		initDelay = 2 * time.Second
		maxDelay  = 1 * time.Minute
	)

	connClose := conn.NotifyClose(make(chan *amqp.Error, 1))

	for {
		select {
		case <-ctx.Done():
			return
		case amqpErr, ok := <-connClose:
			if !ok {
				return
			}
			logger.Warn("rabbitmq: connection lost, reconnecting", zap.Error(amqpErr))
		}

		delay := initDelay
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}

			newConn, err := ConnectRabbitMQ(cfg)
			if err != nil {
				logger.Warn("rabbitmq: reconnect failed, retrying", zap.Duration("delay", delay), zap.Error(err))
				delay *= 2
				if delay > maxDelay {
					delay = maxDelay
				}
				continue
			}

			logger.Info("rabbitmq: connection restored, reinitializing workers")
			onReconnect(newConn)

			connClose = newConn.NotifyClose(make(chan *amqp.Error, 1))
			break
		}
	}
}
