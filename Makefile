.PHONY: help migrate-up migrate-down migrate-force migrate-version migrate-create migrate-legacy-up migrate-legacy-down migrate-legacy-fresh migrate-seed build run docker-up docker-down docker-logs

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Docker commands
docker-up: ## Start development environment (PostgreSQL, Redis, RabbitMQ)
	@echo "Starting development environment..."
	@docker-compose -f docker-compose.dev.yml up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Services started successfully!"
	@echo "PostgreSQL: localhost:5432"
	@echo "Redis: localhost:6379"
	@echo "RabbitMQ: localhost:5672 (Management UI: http://localhost:15672)"

docker-down: ## Stop development environment
	@echo "Stopping development environment..."
	@docker-compose -f docker-compose.dev.yml down

docker-logs: ## View logs from development services
	@docker-compose -f docker-compose.dev.yml logs -f

docker-clean: ## Stop and remove all containers, volumes, and networks
	@echo "Cleaning up development environment..."
	@docker-compose -f docker-compose.dev.yml down -v
	@echo "Cleanup completed!"

# Golang-migrate commands (recommended for production)
migrate-up: ## Run all pending migrations
	@echo "Running migrations..."
	@go run cmd/migrate/main.go up

migrate-down: ## Rollback the last migration
	@echo "Rolling back last migration..."
	@go run cmd/migrate/main.go down

migrate-force: ## Force migration to a specific version (usage: make migrate-force VERSION=1)
	@echo "Forcing migration to version $(VERSION)..."
	@go run cmd/migrate/main.go force $(VERSION)

migrate-version: ## Show current migration version
	@go run cmd/migrate/main.go version

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=create_users_table)
	@echo "Creating migration: $(NAME)"
	@go run cmd/migrate/main.go create $(NAME)

# Legacy GORM migration commands (for development)
migrate-legacy-up: ## Run GORM auto-migrations (development only)
	@echo "Running GORM migrations..."
	@go run cmd/migrate/main.go migrate

migrate-legacy-down: ## Rollback GORM migrations (development only)
	@echo "Rolling back GORM migrations..."
	@go run cmd/migrate/main.go rollback

migrate-legacy-fresh: ## Drop all tables and recreate (development only)
	@echo "Running fresh GORM migrations..."
	@go run cmd/migrate/main.go fresh

# Build and run commands
build: ## Build the application
	@echo "Building application..."
	@go build -o bin/server cmd/server/main.go

run: ## Run the application
	@echo "Running application..."
	@go run cmd/server/main.go

# Development helpers
install-migrate-cli: ## Install golang-migrate CLI tool
	@echo "Installing golang-migrate CLI..."
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

test: ## Run tests
	@echo "Running tests..."
	@go test ./... -v

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf tmp/
