package config

import (
	"fmt"
	"go-gin-clean/pkg/utils"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	JWT        JWTConfig
	OAuth      OAuthConfig
	AES        AESConfig
	RabbitMQ   RabbitMQConfig
	Cloudinary CloudinaryConfig
	Redis      RedisConfig
	Email      EmailConfig
}

type ServerConfig struct {
	AppName     string
	Host        string
	Port        int
	Environment string
	AppUrl      string
	Timeout     int
}

type DatabaseConfig struct {
	Url          string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
	MaxIdleTime  time.Duration
}

type JWTConfig struct {
	JWTIssuer          string
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

type OAuthConfig struct {
	GoogleClientID       string
	GoogleClientSecret   string
	GoogleRedirectURL    string
	GoogleAllowedOrigins []string

	OAuthStateString string
	FrontendURLs     map[string]string
	MobileDeepLinks  map[string]string
	DefaultAppID     string
}

type AESConfig struct {
	Key string
	IV  string
}

type RabbitMQConfig struct {
	Url           string
	Exchange      string
	QueueBaseName string
}

type CloudinaryConfig struct {
	CloudinaryURL string
}

type RedisConfig struct {
	Host       string
	Port       int
	Password   string
	DB         int
	Expiration int
}

type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func Load() (*Config, error) {
	defaultAppID := getEnv("DEFAULT_APP_ID", "default")

	frontendURLs := utils.ParseFrontendURLs(getEnv("FRONTEND_URLS", ""))
	if len(frontendURLs) == 0 {
		frontendURLs = map[string]string{
			defaultAppID: getEnv("FRONTEND_URL", "http://localhost:3120"),
		}
	}

	return &Config{
		Server: ServerConfig{
			AppName:     getEnv("APP_NAME", "Go Gin Clean Template"),
			Host:        getEnv("SERVER_HOST", "localhost"),
			Port:        getEnvAsInt("SERVER_PORT", 3000),
			Environment: getEnv("ENVIRONMENT", "development"),
			AppUrl:      getEnv("FRONTEND_URL", "http://localhost:8080"),
			Timeout:     getEnvAsInt("TIMEOUT", 30),
		},
		Database: DatabaseConfig{
			Url:          getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/dbname?sslmode=disable"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			MaxLifetime:  getEnvAsDuration("DB_MAX_LIFETIME", 30*time.Minute),
			MaxIdleTime:  getEnvAsDuration("DB_MAX_IDLE_TIME", 10*time.Minute),
		},
		JWT: JWTConfig{
			JWTIssuer:          getEnv("JWT_ISSUER", "go-gin-clean"),
			AccessTokenSecret:  getEnv("JWT_ACCESS_SECRET", "your-access-secret-key"),
			RefreshTokenSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key"),
			AccessTokenExpiry:  getEnvAsDuration("JWT_ACCESS_EXPIRY", 1*time.Hour),
			RefreshTokenExpiry: getEnvAsDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
		},
		OAuth: OAuthConfig{
			GoogleClientID:       getEnv("GOOGLE_CLIENT_ID", "your-google-client-id"),
			GoogleClientSecret:   getEnv("GOOGLE_CLIENT_SECRET", "your-google-client-secret"),
			GoogleRedirectURL:    getEnv("GOOGLE_REDIRECT_URL", "http://localhost:3121/api/v1/auth/oauth2/google/callback"),
			GoogleAllowedOrigins: utils.ParseAllowedOrigins(getEnv("GOOGLE_ALLOWED_ORIGINS", "http://localhost:3120")),

			OAuthStateString: getEnv("OAUTH_STATE_STRING", "your-random-state-string"),
			FrontendURLs:     frontendURLs,
			MobileDeepLinks:  utils.ParseFrontendURLs(getEnv("MOBILE_DEEP_LINKS", "")),
			DefaultAppID:     defaultAppID,
		},
		AES: AESConfig{
			Key: getEnv("AES_KEY", "your-aes-encryption-key"),
			IV:  getEnv("AES_IV", "your-aes-initialization-vector"),
		},
		RabbitMQ: RabbitMQConfig{
			Url:           getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			Exchange:      getEnv("RABBITMQ_EXCHANGE", "main_event_bus"),
			QueueBaseName: getEnv("RABBITMQ_QUEUE_BASE_NAME", "main-service"),
		},
		Cloudinary: CloudinaryConfig{
			CloudinaryURL: getEnv("CLOUDINARY_URL", "cloudinary://API_KEY:API_SECRET@CLOUD_NAME"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Email: EmailConfig{
			Host:     getEnv("EMAIL_HOST", "smtp.example.com"),
			Port:     getEnvAsInt("EMAIL_PORT", 587),
			Username: getEnv("EMAIL_USERNAME", ""),
			Password: getEnv("EMAIL_PASSWORD", ""),
			From:     getEnv("EMAIL_FROM", ""),
		},
	}, nil
}

func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Helper
func getEnv(key string, defaultValue string) string {
	if os.Getenv(key) != "" {
		return os.Getenv(key)
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
