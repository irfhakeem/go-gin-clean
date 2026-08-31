package port

import (
	"context"
	"go-gin-clean/internal/domain/entity"
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

type OAuthProvider interface {
	GetGoogleAuthURL(appID string, platform string) string
	HandleGoogleCallback(ctx context.Context, state string, code string) (*entity.User, string, string, error)
	GetFrontendURL(appID string) string
	GetMobileDeepLinkURL(appID string) string
}

type Storage interface {
	UploadFile(ctx context.Context, filename string, size int64, fileHeader multipart.FileHeader, filePath string) (*string, error)
	DeleteFile(ctx context.Context, path string) error
}
type Cache interface {
	Set(ctx context.Context, key string, value any) error
	Get(ctx context.Context, key string, dest any) error
	GetString(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetWithExpiration(ctx context.Context, key string, value any, expiration time.Duration) error
}
