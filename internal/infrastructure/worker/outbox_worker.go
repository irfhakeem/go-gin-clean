package worker

import (
	"context"
	"log"
	"time"

	"go-gin-clean/internal/usecase"
)

const (
	defaultBatchSize    = 10
	defaultPollInterval = 5 * time.Second
	defaultStuckTimeout = 5 * time.Minute
)

type OutboxWorker struct {
	outboxUseCase usecase.OutboxUseCaseInterface
	batchSize     int
	pollInterval  time.Duration
	stuckTimeout  time.Duration
}

func NewOutboxWorker(outboxUseCase usecase.OutboxUseCaseInterface) *OutboxWorker {
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
	log.Println("[OutboxWorker] starting")

	if err := w.outboxUseCase.ResetStuck(ctx, w.stuckTimeout); err != nil {
		log.Printf("[OutboxWorker] ResetStuck on startup error: %v", err)
	}

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	stuckTick := 0

	for {
		select {
		case <-ctx.Done():
			log.Println("[OutboxWorker] shutting down")
			return
		case <-ticker.C:
			if err := w.outboxUseCase.ProcessBatch(ctx, w.batchSize); err != nil {
				log.Printf("[OutboxWorker] ProcessBatch error: %v", err)
			}

			stuckTick++
			if stuckTick >= 10 {
				stuckTick = 0
				if err := w.outboxUseCase.ResetStuck(ctx, w.stuckTimeout); err != nil {
					log.Printf("[OutboxWorker] ResetStuck periodic error: %v", err)
				}
			}
		}
	}
}
