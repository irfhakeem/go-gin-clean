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
	Mailer     MailerConfig
	AES        AESConfig
	RabbitMQ   RabbitMQConfig
	Cloudinary CloudinaryConfig
	Redis      RedisConfig
}

type ServerConfig struct {
	Host        string
	Port        int
	Environment string
	AppUrl      string
	Timeout     int
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	MaxOpenConns int
	MaxIdleConns int
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
	DefaultAppID     string
}

type MailerConfig struct {
	Host     string
	Port     int
	Sender   string
	Auth     string
	Password string
}

type AESConfig struct {
	Key string
	IV  string
}

type RabbitMQConfig struct {
	Host     string
	Port     int
	Username string
	Password string
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

func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Host:        getEnv("SERVER_HOST", "localhost"),
			Port:        getEnvAsInt("SERVER_PORT", 3000),
			Environment: getEnv("ENVIRONMENT", "development"),
			AppUrl:      getEnv("FRONTEND_URL", "http://localhost:8080"),
			Timeout:     getEnvAsInt("TIMEOUT", 30),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnvAsInt("DB_PORT", 5432),
			User:         getEnv("DB_USER", "user"),
			Password:     getEnv("DB_PASSWORD", "password"),
			DBName:       getEnv("DB_NAME", "dbname"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
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
			GoogleRedirectURL:    getEnv("GOOGLE_REDIRECT_URL", "http://localhost:3000/oauth/callback/google"),
			GoogleAllowedOrigins: utils.ParseAllowedOrigins(getEnv("GOOGLE_ALLOWED_ORIGINS", "http://localhost:5000")),

			OAuthStateString: getEnv("OAUTH_STATE_STRING", "your-random-state-string"),
			FrontendURLs:     utils.ParseFrontendURLs(getEnv("FRONTEND_URL", "http://localhost:5000:default")),
			DefaultAppID:     getEnv("DEFAULT_APP_ID", "default"),
		},
		Mailer: MailerConfig{
			Host:     getEnv("MAILER_HOST", "smtp.example.com"),
			Port:     getEnvAsInt("MAILER_PORT", 587),
			Sender:   getEnv("MAILER_SENDER", "Go.Gin.Hexagonal <no-reply@testing.com>"),
			Auth:     getEnv("MAILER_AUTH", "your-authentication-string"),
			Password: getEnv("MAILER_PASSWORD", "your-email-password"),
		},
		AES: AESConfig{
			Key: getEnv("AES_KEY", "your-aes-encryption-key"),
			IV:  getEnv("AES_IV", "your-aes-initialization-vector"),
		},
		RabbitMQ: RabbitMQConfig{
			Host:     getEnv("RABBITMQ_HOST", "localhost"),
			Port:     getEnvAsInt("RABBITMQ_PORT", 5672),
			Username: getEnv("RABBITMQ_USER", "guest"),
			Password: getEnv("RABBITMQ_PASSWORD", "guest"),
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
	}, nil
}

func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d",
		c.Host, c.User, c.Password, c.DBName, c.Port)
}

func (c *RabbitMQConfig) DSN() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", c.Username, c.Password, c.Host, c.Port)
}

func GetAppURL() string {
	return getEnv("FRONTEND_URL", "http://localhost:5000")
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
