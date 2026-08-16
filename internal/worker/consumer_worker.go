package worker

import (
	"context"
	"go-gin-clean/internal/usecase"
	"go-gin-clean/pkg/config"
	"go-gin-clean/pkg/logger"

	mq "go-gin-clean/internal/delivery/mq"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type ConsumerWorker struct {
	conn        *amqp.Connection
	reconnectCh chan *amqp.Connection

	emailUsecase usecase.EmailUseCaseInterface

	cfg *config.RabbitMQConfig
}

func NewConsumerWorker(conn *amqp.Connection, emailUsecase usecase.EmailUseCaseInterface, cfg *config.RabbitMQConfig) *ConsumerWorker {
	return &ConsumerWorker{
		cfg:          cfg,
		conn:         conn,
		reconnectCh:  make(chan *amqp.Connection, 1),
		emailUsecase: emailUsecase,
	}
}

func (w *ConsumerWorker) Reconnect(conn *amqp.Connection) {
	w.reconnectCh <- conn
}

func (w *ConsumerWorker) Run(ctx context.Context) {
	for {
		if err := w.engine(ctx); err != nil {
			logger.Error("consumer worker: failed to start consumers", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			logger.Info("consumer worker: context canceled, shutting down")
			return
		case conn := <-w.reconnectCh:
			w.conn = conn
			logger.Info("consumer worker: reconnected to RabbitMQ, reinitializing consumers")
		}
	}
}

func (w *ConsumerWorker) engine(ctx context.Context) error {
	emailHandler := mq.NewEmailEventHandler(w.emailUsecase)

	consumer := mq.NewConsumer(w.conn, w.cfg.Exchange, map[string]mq.EventHandlerFunc{
		"user.register":       emailHandler.HandleUserVerifyEmail,
		"user.reset_password": emailHandler.HandleUserResetPasswordEmail,
	})

	if err := consumer.Initialize(ctx); err != nil {
		return err
	}

	for service := range mq.MainQueues {
		if err := consumer.Consume(ctx, service); err != nil {
			return err
		}
	}

	logger.Info("consumer worker: all consumers running", zap.String("exchange", w.cfg.Exchange))
	return nil
}
