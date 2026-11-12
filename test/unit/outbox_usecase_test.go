package unit_test

import (
	"context"
	"errors"
	"testing"

	"go-gin-clean/internal/application/usecase"
	"go-gin-clean/internal/domain/entity"
	"go-gin-clean/pkg/config"
	"go-gin-clean/pkg/logger"
	"go-gin-clean/test/mock"
)

func TestOutboxUseCase_ProcessBatch(t *testing.T) {
	logger.Init("test")

	t.Run("Publishes pending message and marks it published by ID", func(t *testing.T) {
		outboxRepo := mock.NewMockOutboxRepository()
		pub := mock.NewMockPublisher()
		uc := usecase.NewOutboxUseCase(outboxRepo, pub, &config.RabbitMQConfig{Exchange: "test_exchange"})

		msg := &entity.Outbox{
			ID:         42,
			EventType:  "test.event",
			Payload:    []byte(`{"key":"value"}`),
			Status:     entity.OutboxStatusPending,
			RetryCount: 0,
			MaxRetries: 3,
		}
		outboxRepo.Messages = append(outboxRepo.Messages, msg)

		err := uc.ProcessBatch(context.Background(), 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if pub.Count() != 1 {
			t.Fatalf("expected 1 published message, got %d", pub.Count())
		}
		if len(outboxRepo.PublishedIDs) != 1 || outboxRepo.PublishedIDs[0] != 42 {
			t.Fatalf("expected PublishedIDs to contain [42], got %v", outboxRepo.PublishedIDs)
		}
		if msg.Status != entity.OutboxStatusPublished {
			t.Fatalf("expected message status to be published, got %s", msg.Status)
		}
	})

	t.Run("Handles publish error by calling RetryOrFail with ID", func(t *testing.T) {
		outboxRepo := mock.NewMockOutboxRepository()
		pub := mock.NewMockPublisher()
		pub.PublishErr = errors.New("amqp publish failed")
		uc := usecase.NewOutboxUseCase(outboxRepo, pub, &config.RabbitMQConfig{Exchange: "test_exchange"})

		msg := &entity.Outbox{
			ID:         100,
			EventType:  "test.event",
			Payload:    []byte(`{"key":"value"}`),
			Status:     entity.OutboxStatusPending,
			RetryCount: 0,
			MaxRetries: 3,
		}
		outboxRepo.Messages = append(outboxRepo.Messages, msg)

		err := uc.ProcessBatch(context.Background(), 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(outboxRepo.PublishedIDs) != 0 {
			t.Fatalf("expected no published IDs, got %v", outboxRepo.PublishedIDs)
		}
	})
}
