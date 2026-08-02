package usecase

import (
	"context"
	"encoding/json"
	"time"

	"go-gin-clean/internal/entity"
	"go-gin-clean/internal/gateway/messaging"
	"go-gin-clean/internal/repository"
	"go-gin-clean/pkg/logger"

	"go.uber.org/zap"
)

type OutboxUseCaseInterface interface {
	SaveOutboxMessage(ctx context.Context, aggregateType, aggregateID, eventType string, payload any) error
	ProcessBatch(ctx context.Context, limit int) error
	ResetStuck(ctx context.Context, stuckDuration time.Duration) error
}

type OutboxUseCase struct {
	outboxRepo repository.OutboxRepositoryInterface
	publisher  messaging.PublisherServiceInterface
	exchange   string
}

func NewOutboxUseCase(
	outboxRepo repository.OutboxRepositoryInterface,
	publisher messaging.PublisherServiceInterface,
	exchange string,
) OutboxUseCaseInterface {
	return &OutboxUseCase{
		outboxRepo: outboxRepo,
		publisher:  publisher,
		exchange:   exchange,
	}
}

func (u *OutboxUseCase) ProcessBatch(ctx context.Context, limit int) error {
	messages, err := u.outboxRepo.FetchAndLockPending(ctx, limit)
	if err != nil {
		return err
	}

	for _, msg := range messages {
		if pubErr := u.publisher.Publish(msg.EventType, msg.Payload); pubErr != nil {
			logger.Error("outbox publish failed",
				zap.Int64("pkid", msg.PKID),
				zap.String("event", msg.EventType),
				zap.Error(pubErr),
			)
			if retryErr := u.outboxRepo.RetryOrFail(ctx, msg.PKID, pubErr.Error()); retryErr != nil {
				logger.Error("outbox retry failed",
					zap.Int64("pkid", msg.PKID),
					zap.Error(retryErr),
				)
			}
		} else {
			if markErr := u.outboxRepo.MarkPublished(ctx, msg.PKID); markErr != nil {
				logger.Error("outbox mark published failed",
					zap.Int64("pkid", msg.PKID),
					zap.Error(markErr),
				)
			}
		}
	}

	return nil
}

func (u *OutboxUseCase) ResetStuck(ctx context.Context, stuckDuration time.Duration) error {
	return u.outboxRepo.ResetStuck(ctx, stuckDuration)
}

func (u *OutboxUseCase) SaveOutboxMessage(ctx context.Context, aggregateType, aggregateID, eventType string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	msg := &entity.Outbox{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Exchange:      u.exchange,
		Payload:       body,
		Status:        entity.OutboxStatusPending,
	}
	return u.outboxRepo.Create(ctx, msg)
}
