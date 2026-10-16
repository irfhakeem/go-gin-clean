package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	JWT        JWTConfig
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
	cfg := &Config{
		Server: ServerConfig{
			AppName:     getEnv("APP_NAME", "Go Gin Clean Template"),
			Host:        getEnv("SERVER_HOST", "localhost"),
			Port:        getEnvAsInt("SERVER_PORT", 3000),
			Environment: getEnv("ENVIRONMENT", "development"),
			AppUrl:      getEnv("FRONTEND_URL", "http://localhost:8080"),
			Timeout:     getEnvAsInt("TIMEOUT", 30),
		},
		Database: DatabaseConfig{
			Url:          getEnv("DB_URL", "postgres://user:password@localhost:5432/dbname?sslmode=disable"),
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
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.Environment == "production" {
		insecurePlaceholders := map[string]string{
			"JWT_ACCESS_SECRET":  c.JWT.AccessTokenSecret,
			"JWT_REFRESH_SECRET": c.JWT.RefreshTokenSecret,
			"AES_KEY":            c.AES.Key,
		}

		defaults := []string{
			"your-access-secret-key",
			"your-refresh-secret-key",
			"your-aes-encryption-key",
			"secret",
			"password",
		}

		for name, val := range insecurePlaceholders {
			if len(val) < 16 {
				return fmt.Errorf("production security error: %s must be at least 16 characters", name)
			}
			for _, d := range defaults {
				if val == d {
					return fmt.Errorf("production security error: %s cannot use default placeholder %q", name, d)
				}
			}
		}
	}
	return nil
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
