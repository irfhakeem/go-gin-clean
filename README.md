# Go Gin Clean Architecture

A production-ready Go web application implementing Clean Architecture principles with proper separation of concerns, dependency injection, and modern best practices.

## 🏗️ Architecture Overview

This project follows **Clean Architecture** principles with a pragmatic layered approach:

```
┌──────────────────────────────────────────────────────────┐
│                   External Interfaces                     │
│     HTTP Clients, PostgreSQL, Redis, RabbitMQ, etc.      │
└──────────────────────────────────────────────────────────┘
                            ▲
                            │
┌──────────────────────────────────────────────────────────┐
│                    Delivery Layer                         │
│         HTTP Handlers, Middleware, Routes                 │
│            (Presents data to external world)              │
└──────────────────────────────────────────────────────────┘
                            ▲
                            │
┌──────────────────────────────────────────────────────────┐
│                   Use Case Layer                          │
│         Application Business Logic & Orchestration        │
│    (Coordinates entities, repositories, and gateways)     │
└──────────────────────────────────────────────────────────┘
                            ▲
                            │
┌──────────────────────────┬───────────────────────────────┐
│     Repository Layer     │      Gateway Layer            │
│   (Database Access)      │  (Security, Media, Cache,     │
│   • User Repository      │   Messaging Services)         │
│   • Token Repository     │  • JWT, Bcrypt, AES, OAuth    │
│                          │  • Cloudinary, Local Storage  │
│                          │  • Redis, RabbitMQ            │
└──────────────────────────┴───────────────────────────────┘
                            ▲
                            │
┌──────────────────────────────────────────────────────────┐
│                    Entity Layer                           │
│        Domain Models & Business Rules (Pure Go)           │
│              • User    • RefreshToken                     │
└──────────────────────────────────────────────────────────┘
```

**Dependency Rule**: Inner layers know nothing about outer layers. Dependencies point inward.

## 📁 Project Structure

```
go-gin-clean/
├── cmd/                                    # Application entrypoints
│   ├── server/main.go                     # HTTP server (main entry)
│   └── migrate/main.go                    # Database migration CLI
│
├── internal/                              # Private application code
│   ├── delivery/                          # 📤 Delivery Layer (Presentation)
│   │   └── http/                          # HTTP transport
│   │       ├── middleware/                # Auth, CORS, rate limiting
│   │       ├── response/                  # Standardized API responses
│   │       ├── route/                     # Route registration
│   │       │   └── route.go               # All API routes defined here
│   │       ├── user_handler.go            # User HTTP handlers
│   │       └── oauth_handler.go           # OAuth HTTP handlers
│   │
│   ├── usecase/                           # 🎯 Use Case Layer (Business Logic)
│   │   └── user_usecase.go                # User business logic orchestration
│   │
│   ├── repository/                        # 💾 Repository Layer (Data Access)
│   │   ├── repository.go                  # Base repository interface
│   │   ├── user_repository.go             # User data operations (GORM)
│   │   └── refresh_token_repository.go    # Token persistence
│   │
│   ├── gateway/                           # 🌐 Gateway Layer (External Services)
│   │   ├── security/                      # Security services
│   │   │   ├── jwt_service.go             # JWT generation & validation
│   │   │   ├── bcrypt_service.go          # Password hashing
│   │   │   ├── aes_service.go             # AES encryption/decryption
│   │   │   └── oauth_service.go           # Google OAuth integration
│   │   ├── media/                         # File storage services
│   │   │   ├── localstorage_service.go    # Local file system storage
│   │   │   └── cloudinary_service.go      # Cloudinary cloud storage
│   │   ├── cache/                         # Caching services
│   │   │   └── redis.go                   # Redis cache operations
│   │   └── messaging/                     # Async messaging
│   │       ├── publisher.go               # RabbitMQ base publisher
│   │       └── user_publisher.go          # User event publisher
│   │
│   ├── entity/                            # 🏛️ Entity Layer (Domain Models)
│   │   ├── user.go                        # User entity with business rules
│   │   ├── refresh_token.go               # RefreshToken entity
│   │   └── audit.go                       # Audit fields (created/updated)
│   │
│   ├── model/                             # 📋 DTOs & Transfer Objects
│   │   ├── user_model.go                  # User request/response DTOs
│   │   ├── oauth_model.go                 # OAuth DTOs
│   │   ├── claims_model.go                # JWT claims
│   │   ├── user_event.go                  # Event payloads for messaging
│   │   └── pagination.go                  # Pagination utilities
│   │
│   └── infrastructure/                    # 🔧 Infrastructure (Dependency Injection)
│       └── container.go                   # IoC container for wiring dependencies
│
├── pkg/                                   # Public shared packages
│   ├── config/                            # Configuration management
│   │   └── config.go                      # Environment variable loader
│   ├── errors/                            # Application error definitions
│   │   └── errors.go                      # Centralized error messages
│   └── utils/                             # Utility functions
│       ├── string_utils.go                # String helpers
│       └── number_utils.go                # Number helpers
│
├── migrations/                            # 📊 Database migrations (golang-migrate)
│   ├── 000001_create_enums.up.sql        # Create enum types
│   ├── 000001_create_enums.down.sql
│   ├── 000002_create_users_table.up.sql  # Create users table
│   ├── 000002_create_users_table.down.sql
│   ├── 000003_create_refresh_tokens_table.up.sql
│   └── 000003_create_refresh_tokens_table.down.sql
│
├── assets/                                # Static assets & uploaded files
├── .env.example                           # Environment variables template
├── .air.toml                              # Air hot reload configuration
├── Dockerfile                             # Production Docker image
├── Makefile                               # Development commands
├── go.mod                                 # Go module definition
└── go.sum                                 # Dependency checksums
```

### Key Architectural Decisions

- **Separation of Concerns**: Each layer has a single responsibility
- **Dependency Inversion**: Inner layers define interfaces, outer layers implement them
- **No Circular Dependencies**: Dependencies flow inward (Entity ← Repository/Gateway ← UseCase ← Delivery)
- **Testability**: Business logic isolated from frameworks and external services
- **Flexibility**: Easy to swap implementations (e.g., switch from local storage to S3)

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** (1.24.3 used in this project)
- **PostgreSQL 12+** (primary database)
- **Redis** (optional, for caching)
- **RabbitMQ** (optional, for async email notifications)
- **Docker** (optional, for running dependencies)

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/yourusername/go-gin-clean.git
   cd go-gin-clean
   ```

2. **Install Go dependencies**

   ```bash
   go mod download
   ```

3. **Setup environment variables**

   Copy `.env.example` to `.env` and configure your settings:

   ```bash
   cp .env.example .env
   ```

   **Key environment variables:**

   ```env
   # Server Configuration
   SERVER_HOST=localhost
   SERVER_PORT=3000
   ENVIRONMENT=development
   FRONTEND_URL=http://localhost:3120
   TIMEOUT=30

   # Database
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=your_password_here
   DB_NAME=go_clean_architecture
   DB_MAX_OPEN_CONNS=100
   DB_MAX_IDLE_CONNS=10

   # JWT Authentication
   JWT_ISSUER=your-app-name
   JWT_ACCESS_SECRET=your-super-secret-access-key-change-this-in-production
   JWT_REFRESH_SECRET=your-super-secret-refresh-key-change-this-in-production
   JWT_ACCESS_EXPIRY=15m
   JWT_REFRESH_EXPIRY=168h

   # AES Encryption (for tokens & sensitive data)
   AES_KEY=your-32-character-secret-key
   AES_IV=your-16-character-init-vector

   # Google OAuth 2.0
   GOOGLE_CLIENT_ID=your-google-client-id
   GOOGLE_CLIENT_SECRET=your-google-client-secret
   GOOGLE_REDIRECT_URL=http://localhost:3120/callback
   GOOGLE_ALLOWED_ORIGINS=http://localhost:3120
   OAUTH_STATE_STRING=random-secure-state-string

   # Cloudinary (for cloud file storage)
   CLOUDINARY_URL=cloudinary://api_key:api_secret@cloud_name

   # Redis (optional - for caching)
   REDIS_HOST=localhost
   REDIS_PORT=6379
   REDIS_PASSWORD=
   REDIS_DB=0
   REDIS_EXPIRATION=604800

   # RabbitMQ (optional - for async messaging)
   RABBITMQ_HOST=localhost
   RABBITMQ_PORT=5672
   RABBITMQ_USER=guest
   RABBITMQ_PASSWORD=guest
   ```

4. **Start development dependencies (PostgreSQL, Redis, RabbitMQ)**

   Using Docker Compose:

   ```bash
   make docker-up
   ```

   Or start PostgreSQL manually and skip optional services.

5. **Run database migrations**

   Choose one of two migration approaches:

   **Option A: golang-migrate (Recommended for production)**
   ```bash
   make migrate-up
   ```

   **Option B: GORM Auto-migrate (Quick for development)**
   ```bash
   make migrate-legacy-up
   ```

6. **Start the application**

   ```bash
   make run
   # or
   go run cmd/server/main.go
   ```

The server will start on `http://localhost:3000`

### Development with Hot Reload

Install [Air](https://github.com/cosmtrek/air) for hot reloading:

```bash
go install github.com/cosmtrek/air@latest
air
```

## 📚 API Documentation

### Health Check

- `GET /health` - Server health status

### Authentication

- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh-token` - Refresh access token
- `POST /api/v1/auth/verify-email` - Verify email address
- `POST /api/v1/auth/send-reset-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password with token
- `POST /api/v1/auth/resend-verification` - Resend verification email

### OAuth 2.0

- `POST /api/v1/auth/oauth2/url` - Get OAuth provider login URL
- `GET /api/v1/auth/oauth2/:provider/callback` - OAuth callback handler (Google)

### Profile (Authenticated)

- `GET /api/v1/profile` - Get current user profile
- `PUT /api/v1/profile` - Update profile (name, avatar, gender)
- `PUT /api/v1/profile/change-password` - Change password
- `POST /api/v1/profile/logout` - Logout (revoke tokens)

### User Management (Authenticated)

- `GET /api/v1/users` - Get all users (paginated, searchable)
- `GET /api/v1/users/:code` - Get user by code
- `POST /api/v1/users` - Create new user
- `PUT /api/v1/users/:code` - Update user
- `PUT /api/v1/users/:code/change-status` - Change user active status
- `DELETE /api/v1/users/:code` - Delete user

### Authentication Header

Include the access token in protected requests:

```
Authorization: Bearer <access_token>
```

## 🔧 Available Commands

This project uses a `Makefile` for common operations. Run `make help` to see all available commands.

### Docker Commands

```bash
make docker-up           # Start PostgreSQL, Redis, RabbitMQ in Docker
make docker-down         # Stop all Docker services
make docker-logs         # View Docker service logs
make docker-clean        # Remove all containers, volumes, and networks
```

### Database Migration Commands

**Golang-migrate (Production-ready SQL migrations)**

```bash
make migrate-up          # Run all pending migrations
make migrate-down        # Rollback last migration
make migrate-version     # Show current migration version
make migrate-force VERSION=1  # Force migration to specific version
make migrate-create NAME=add_users  # Create new migration files

# Raw commands
go run cmd/migrate/main.go up
go run cmd/migrate/main.go down
go run cmd/migrate/main.go version
go run cmd/migrate/main.go create <migration_name>
```

**GORM Auto-migrate (Development only)**

```bash
make migrate-legacy-up    # Run GORM auto-migrations
make migrate-legacy-down  # Drop all tables
make migrate-legacy-fresh # Drop and recreate tables
```

### Development Commands

```bash
make run                 # Start the application
make build               # Build binary to bin/server
make test                # Run all tests
make clean               # Remove build artifacts

# Direct Go commands
go run cmd/server/main.go          # Start server
go build -o bin/server cmd/server/main.go   # Build server
go test ./...                      # Run tests
go test -v ./...                   # Run tests (verbose)
go test -cover ./...               # Run tests with coverage
go vet ./...                       # Check for issues
go fmt ./...                       # Format code
go mod tidy                        # Clean up dependencies
```

### Production Build & Deployment

```bash
# Build optimized binary
go build -ldflags="-s -w" -o bin/server cmd/server/main.go

# Build Docker image
docker build -t go-gin-clean:latest .

# Run Docker container
docker run -p 3000:3000 --env-file .env go-gin-clean:latest
```

### Useful Development Tools

```bash
# Install Air for hot reload
go install github.com/cosmtrek/air@latest

# Install golang-migrate CLI
make install-migrate-cli

# Run with hot reload
air
```
