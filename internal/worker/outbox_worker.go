package worker

import (
	"context"
	"time"

	"go-gin-clean/internal/application/usecase"
	"go-gin-clean/pkg/logger"

	"go.uber.org/zap"
)

const (
	defaultBatchSize    = 10
	defaultPollInterval = 5 * time.Second
	defaultStuckTimeout = 5 * time.Minute
)

type OutboxWorker struct {
	outboxUseCase usecase.OutboxUseCase
	batchSize     int
	pollInterval  time.Duration
	stuckTimeout  time.Duration
}

func NewOutboxWorker(outboxUseCase usecase.OutboxUseCase) *OutboxWorker {
	return &OutboxWorker{
		outboxUseCase: outboxUseCase,
		batchSize:     defaultBatchSize,
		pollInterval:  defaultPollInterval,
		stuckTimeout:  defaultStuckTimeout,
	}
}

func (w *OutboxWorker) WithBatchSize(n int) *OutboxWorker {
	w.batchSize = n
	return w
}

func (w *OutboxWorker) WithPollInterval(d time.Duration) *OutboxWorker {
	w.pollInterval = d
	return w
}

func (w *OutboxWorker) Run(ctx context.Context) {
	logger.Info("outbox worker starting")

	if err := w.outboxUseCase.ResetStuck(ctx, w.stuckTimeout); err != nil {
		logger.Error("outbox worker reset stuck on startup", zap.Error(err))
	}

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	stuckTick := 0

	for {
		select {
		case <-ctx.Done():
			logger.Info("outbox worker shutting down")
			return
		case <-ticker.C:
			if err := w.outboxUseCase.ProcessBatch(ctx, w.batchSize); err != nil {
				logger.Error("outbox worker process batch", zap.Error(err))
			}

			stuckTick++
			if stuckTick >= 10 {
				stuckTick = 0
				if err := w.outboxUseCase.ResetStuck(ctx, w.stuckTimeout); err != nil {
					logger.Error("outbox worker reset stuck periodic", zap.Error(err))
				}
			}
		}
	}
}
