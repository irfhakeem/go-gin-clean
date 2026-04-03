package entity

import "time"

type OutboxStatus string

const (
	OutboxStatusPending    OutboxStatus = "pending"
	OutboxStatusProcessing OutboxStatus = "processing"
	OutboxStatusPublished  OutboxStatus = "published"
	OutboxStatusFailed     OutboxStatus = "failed"
)

func (s OutboxStatus) String() string {
	return string(s)
}

type Outbox struct {
	PKID          int64        `gorm:"column:pkid;primaryKey;autoIncrement"`
	AggregateType string       `gorm:"column:aggregate_type;not null"`
	AggregateID   string       `gorm:"column:aggregate_id;not null"`
	EventType     string       `gorm:"column:event_type;not null"`
	Exchange      string       `gorm:"column:exchange;not null"`
	Payload       []byte       `gorm:"column:payload;type:jsonb;not null"`
	Status        OutboxStatus `gorm:"column:status;type:outbox_status;default:'pending';not null"`
	RetryCount    int          `gorm:"column:retry_count;default:0;not null"`
	MaxRetries    int          `gorm:"column:max_retries;default:3;not null"`
	LastError     *string      `gorm:"column:last_error;default:null"`
	CreatedAt     time.Time    `gorm:"column:created_at;type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time    `gorm:"column:updated_at;type:timestamp;default:CURRENT_TIMESTAMP"`
	PublishedAt   *time.Time   `gorm:"column:published_at;default:null"`
}

func (Outbox) TableName() string {
	return "outbox_messages"
}
