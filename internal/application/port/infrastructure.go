package port

import (
	"context"
	"mime/multipart"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(routingKey string, body []byte) error
	UpdateConnection(conn *amqp.Connection) error
}

type Consumer interface {
	Initialize(ctx context.Context) error
	Consume(ctx context.Context, queueName string) error
}

type Mailer interface {
	SendEmail(to string, subject string, body string) error
	LoadTemplate(templateName string, data any) (string, error)
}

type Storage interface {
	UploadFile(ctx context.Context, filename string, size int64, fileHeader multipart.FileHeader, filePath string) (*string, error)
	DeleteFile(ctx context.Context, path string) error
}

type Cache interface {
	Get(ctx context.Context, key string, dest any) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	SetWithExpiration(ctx context.Context, key string, value any, expiration time.Duration) error
}
