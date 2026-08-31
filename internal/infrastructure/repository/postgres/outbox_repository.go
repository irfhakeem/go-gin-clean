package postgres

import (
	"context"
	"time"

	"go-gin-clean/internal/domain/entity"

	"gorm.io/gorm"
)

type PostgresOutboxRepository struct {
	db *gorm.DB
}

func NewPostgresOutboxRepository(db *gorm.DB) *PostgresOutboxRepository {
	return &PostgresOutboxRepository{db: db}
}

func (r *PostgresOutboxRepository) Create(ctx context.Context, msg *entity.Outbox) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *PostgresOutboxRepository) FetchAndLockPending(ctx context.Context, limit int) ([]*entity.Outbox, error) {
	var messages []*entity.Outbox

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(`
			SELECT * FROM outbox_messages
			WHERE status = 'pending'
			  AND retry_count < max_retries
			ORDER BY created_at ASC
			LIMIT ?
			FOR UPDATE SKIP LOCKED
		`, limit).Scan(&messages).Error; err != nil {
			return err
		}

		if len(messages) == 0 {
			return nil
		}

		ids := make([]int64, len(messages))
		for i, m := range messages {
			ids[i] = m.PKID
		}

		return tx.Exec(`
			UPDATE outbox_messages
			SET status = 'processing', updated_at = NOW()
			WHERE pkid IN ?
		`, ids).Error
	})

	return messages, err
}

func (r *PostgresOutboxRepository) MarkPublished(ctx context.Context, pkid int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Exec(`
		UPDATE outbox_messages
		SET status = 'published', published_at = ?, updated_at = NOW()
		WHERE pkid = ?
	`, now, pkid).Error
}

func (r *PostgresOutboxRepository) RetryOrFail(ctx context.Context, pkid int64, errMsg string) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE outbox_messages
		SET
			retry_count = retry_count + 1,
			last_error  = ?,
			status = CASE
				WHEN retry_count + 1 >= max_retries THEN 'failed'::outbox_status
				ELSE 'pending'::outbox_status
			END,
			updated_at = NOW()
		WHERE pkid = ?
	`, errMsg, pkid).Error
}

func (r *PostgresOutboxRepository) ResetStuck(ctx context.Context, stuckDuration time.Duration) error {
	cutoff := time.Now().Add(-stuckDuration)
	return r.db.WithContext(ctx).Exec(`
		UPDATE outbox_messages
		SET status = 'pending', updated_at = NOW()
		WHERE status = 'processing'
		  AND updated_at < ?
	`, cutoff).Error
}
