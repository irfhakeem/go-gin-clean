# Go Gin Clean Architecture

A production-ready Go web application built with Go 1.24, Gin framework, and Clean Architecture principles. Features include JWT authentication, OAuth 2.0, UUID-based identifiers, Redis caching, file uploads, and async messaging.

## ✨ Key Features

- 🏗️ **Clean Architecture** - Clear separation of concerns with layered design
- 🔐 **JWT Authentication** - Secure access & refresh token system
- 🌐 **OAuth 2.0** - Google login integration
- 🆔 **UUID Identifiers** - Secure, globally unique user IDs
- 📝 **User Management** - Complete CRUD operations with code-based lookup
- 📧 **Email Verification** - Secure email verification flow
- 🔑 **Password Reset** - Token-based password recovery
- 📤 **File Upload** - Avatar uploads with validation (JPG, JPEG, PNG)
- 💾 **Redis Caching** - Fast user data retrieval with cache invalidation
- 📨 **Async Messaging** - Event-driven architecture with RabbitMQ
- 🔒 **Security** - Bcrypt passwords, AES encryption, CORS protection
- 🗃️ **Database Migrations** - Version-controlled schema changes
- 🚀 **Hot Reload** - Fast development with Air

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
│   ├── 000003_create_refresh_tokens_table.down.sql
│   ├── 000004_convert_pkid_to_uuid.up.sql    # Convert IDs to UUID
│   └── 000004_convert_pkid_to_uuid.down.sql
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

- **Go 1.24+** (Go 1.24.3 recommended)
- **PostgreSQL 12+** with UUID extension support
- **Redis 6+** (optional, for caching)
- **RabbitMQ 3.8+** (optional, for async messaging)
- **Docker & Docker Compose** (optional, for running dependencies)

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

   **Production Migrations (Recommended)**

   Use SQL-based migrations for production environments:

   ```bash
   make migrate-up          # Apply all pending migrations
   ```

   **Development Auto-Migration (Quick Setup)**

   Use GORM auto-migrate for rapid development:

   ```bash
   make migrate-legacy-up   # Auto-generate schema from Go models
   ```

   > **Note**:
   >
   > - `migrate-up` uses versioned SQL migrations (recommended for production)
   > - `migrate-legacy-up` uses GORM auto-migrate (quick for development only)
   > - Production should always use SQL migrations for version control

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

### Base URL

```
http://localhost:3000
```

### Response Format

All API responses follow a consistent structure:

**Success Response:**

```json
{
  "success": true,
  "message": "Operation successful",
  "data": {
    /* response data */
  }
}
```

**Error Response:**

```json
{
  "success": false,
  "message": "Error occurred",
  "error": "Detailed error message"
}
```

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

### Docker Commands

```bash
make docker-up           # Start PostgreSQL, Redis, RabbitMQ containers
make docker-down         # Stop all Docker services
make docker-logs         # View Docker service logs
make docker-clean        # Remove containers, volumes, and networks
```

### Database Migration Commands

#### Production Migrations (SQL-based)

Use these commands for production environments with version-controlled SQL migrations:

```bash
make migrate-up              # Apply all pending migrations
make migrate-down            # Rollback the last migration
make migrate-version         # Show current migration version
make migrate-force VERSION=4 # Force set migration version (use with caution)
make migrate-create NAME=add_new_feature  # Create new migration files
```

**Direct commands:**

```bash
go run cmd/migrate/main.go up          # Apply migrations
go run cmd/migrate/main.go down        # Rollback migration
go run cmd/migrate/main.go version     # Check version
go run cmd/migrate/main.go create <name>  # Create migration
```

#### Development Auto-Migration (GORM-based)

Use these commands for rapid development and testing:

```bash
make migrate-legacy-up       # Auto-generate schema from Go models
make migrate-legacy-down     # Drop all tables
make migrate-legacy-fresh    # Drop and recreate all tables
```

> **Important Notes:**
>
> - **Production**: Always use SQL migrations (`migrate-up/down`) for production databases
> - **Development**: Use auto-migration (`migrate-legacy-up`) for quick local setup
> - SQL migrations provide version control and rollback capabilities
> - Auto-migration is destructive and should never be used in production

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

# Format code
go fmt ./...

# Lint code
go vet ./...

# Check for vulnerabilities
go list -json -m all | nancy sleuth
```

## 🛠️ Technology Stack

### Core

- **Go 1.24** - Programming language
- **Gin** - HTTP web framework
- **GORM** - ORM library

### Database

- **PostgreSQL** - Primary database
- **UUID v4** - Unique identifiers (via `pgcrypto`)
- **golang-migrate** - Database migrations

### Authentication & Security

- **JWT** - JSON Web Tokens (access & refresh tokens)
- **OAuth 2.0** - Google authentication
- **Bcrypt** - Password hashing
- **AES-256** - Symmetric encryption for sensitive data

### Caching & Messaging

- **Redis** - Response caching with TTL
- **RabbitMQ** - Async event publishing (email notifications)

### File Storage

- **Local Storage** - File system uploads
- **Cloudinary** - Cloud media storage (optional)

### Development Tools

- **Air** - Hot reload
- **Docker Compose** - Local development environment
- **Make** - Task automation

## 📊 Database Schema

### Users Table

- `id` (UUID) - Primary key
- `code` (VARCHAR) - Unique user code (auto-generated)
- `name` (VARCHAR) - User full name
- `email` (VARCHAR) - Unique email address
- `password` (VARCHAR) - Bcrypt hashed password
- `avatar` (VARCHAR) - Avatar file path/URL
- `gender` (ENUM) - Male, Female, Other
- `role` (ENUM) - Admin, User
- `is_active` (BOOLEAN) - Account status
- `is_verified` (BOOLEAN) - Email verification status
- `oauth_provider` (VARCHAR) - OAuth provider (e.g., "google")
- `oauth_id` (VARCHAR) - OAuth user ID
- Audit fields: `created_at`, `updated_at`, `deleted_at`, `is_deleted`

### Refresh Tokens Table

- `id` (UUID) - Primary key
- `user_id` (UUID) - Foreign key to users
- `token` (VARCHAR) - Encrypted refresh token
- `expiry_at` (TIMESTAMP) - Token expiration
- `is_revoked` (BOOLEAN) - Revocation status
- Audit fields: `created_at`, `updated_at`, `deleted_at`, `is_deleted`

## 🔐 Security Features

- **JWT Authentication** - Stateless authentication with short-lived access tokens
- **Refresh Tokens** - Long-lived tokens for obtaining new access tokens
- **Password Hashing** - Bcrypt with cost factor 10
- **AES Encryption** - URL-safe encryption for email verification & password reset tokens
- **OAuth 2.0** - Secure third-party authentication
- **CORS Protection** - Configurable allowed origins
- **Rate Limiting** - Prevent abuse (can be added via middleware)
- **Input Validation** - Request validation using Gin binding
- **Error Masking** - User-friendly errors without exposing internals
- **UUID IDs** - No sequential ID enumeration attacks

## 🚀 Deployment

### Environment Variables

Ensure all required environment variables are set:

```bash
# Copy and configure
cp .env.example .env
nano .env
```

### Build for Production

```bash
# Build optimized binary
make build

# Or with custom flags
go build -ldflags="-s -w" -o bin/server cmd/server/main.go
```

### Docker Deployment

```bash
# Build image
docker build -t go-gin-clean:latest .

# Run container
docker run -d \
  --name go-gin-clean \
  -p 3000:3000 \
  --env-file .env \
  go-gin-clean:latest
```

### Database Migration in Production

```bash
# Apply migrations before deploying new version
make migrate-up

# Or using the binary
./bin/migrate up
```

## 📝 Common Tasks

### Creating a New Migration

```bash
make migrate-create NAME=add_user_preferences

# This creates:
# migrations/000005_add_user_preferences.up.sql
# migrations/000005_add_user_preferences.down.sql
```

### Rollback Last Migration

```bash
make migrate-down
```

### Invalidate User Cache

Cache is automatically invalidated on:

- User creation
- User update
- User deletion
- Status change

Manual cache invalidation happens in the usecase layer.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
