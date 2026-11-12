package inbound

import (
	"context"
	"time"
)

type OutboxUseCase interface {
	SaveOutboxMessage(ctx context.Context, aggregateType, aggregateID, eventType string, payload any) error
	ProcessBatch(ctx context.Context, limit int) error
	ResetStuck(ctx context.Context, stuckDuration time.Duration) error
}
