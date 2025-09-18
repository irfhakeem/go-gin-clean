# Go Gin Clean Architecture

A complete Go web application implementing Clean Architecture principles with proper separation of concerns and dependency inversion.

## 🏗️ Architecture Overview

This project follows **Clean Architecture** principles, organizing code into clear layers with proper dependency direction:

```
┌─────────────────────────────────────────┐
│            External Interfaces          │
│        (Database, Web, CLI, etc.)       │
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│               Adapters                  │
│         (Controllers, Gateways,         │
│          Presenters, etc.)              │
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│            Use Cases                    │
│        (Application Business            │
│             Rules)                      │
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│              Entities                   │
│         (Enterprise Business            │
│              Rules)                     │
└─────────────────────────────────────────┘
```

## 📁 Project Structure

```
go-gin-clean/
├── cmd/                          # Application entrypoints
│   ├── server/main.go           # HTTP server
│   └── migrate/main.go          # Database migrations
├── internal/                    # Private application code
│   ├── core/                    # Core business logic (innermost layer)
│   │   ├── domain/              # Enterprise business rules
│   │   │   ├── entities/        # Business entities (User, RefreshToken, Audit)
│   │   │   ├── enums/           # Enumerations (Gender)
│   │   │   └── errors/          # Domain errors
│   │   ├── dto/                 # Data Transfer Objects
│   │   ├── ports/               # Interfaces (contracts)
│   │   └── usecases/            # Application business rules
│   │       ├── user_usecase.go  # User business logic
│   │       └── email_usecase.go # Email business logic
│   ├── adapters/                # Adapters for external interfaces
│   │   ├── primary/http/        # HTTP layer
│   │   │   ├── handlers/        # HTTP handlers
│   │   │   ├── messages/        # Response messages
│   │   │   ├── response/        # Response utilities
│   │   │   ├── middleware.go    # Authentication middleware
│   │   │   └── routes.go        # Route definitions
│   │   └── secondary/           # External service implementations
│   │       ├── database/        # Database repositories
│   │       ├── security/        # JWT, Bcrypt, AES services
│   │       ├── mailer/          # SMTP email service
│   │       └── media/           # Local storage service
│   └── infrastructure/          # Infrastructure concerns
│       └── container.go         # Dependency injection
└── pkg/                         # Public libraries
    ├── config/                  # Configuration management
    └── utils/                   # Utility functions
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL
- Git

### Setup

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd go-gin-clean
   ```

2. **Install dependencies**

   ```bash
   go mod download
   ```

3. **Environment setup**
   Copy `.env.example` to `.env` and configure:

   ```env
   # Server
   SERVER_HOST=localhost
   SERVER_PORT=8080
   ENVIRONMENT=development
   APP_FE_URL=
   TIMEOUT=30

   # Database
   DB_HOST=localhost
   DB_PORT=5432
   DB_USERNAME=postgres
   DB_PASSWORD=your_password
   DB_NAME=go_gin_clean
   DB_MAX_IDLE_CONNS=25
   DB_MAX_OPEN_CONNS=5

   # JWT
   JWT_ISSUER=go-gin-clean
   JWT_ACCESS_SECRET=your-access-secret-key
   JWT_REFRESH_SECRET=your-refresh-secret-key
   JWT_ACCESS_EXPIRY=1h
   JWT_REFRESH_EXPIRY=168h

   # AES Encryption
   AES_KEY=your-32-character-encryption-key
   AES_IV=your-16-character-iv-key

   # SMTP Email (optional)
   MAILER_HOST=smtp.gmail.com
   MAILER_PORT=587
   MAILER_SENDER=your-email@gmail.com
   MAILER_AUTH=your-email@gmail.com
   MAILER_PASSWORD=your-app-password
   ```

4. **Database migration**

   ```bash
   go run cmd/migrate/main.go migrate
   ```

5. **Start the server**
   ```bash
   go run cmd/server/main.go
   ```

The server will start on `http://localhost:8080`

## 📚 API Documentation

### Health Check

- `GET /health` - Server health status

### Authentication (Public Routes)

- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login (sets refresh token in cookie)
- `POST /api/v1/auth/refresh-token` - Refresh access token
- `POST /api/v1/auth/verify-email` - Email verification
- `POST /api/v1/auth/send-verify-email` - Send verification email
- `POST /api/v1/auth/send-reset-password` - Send reset password email
- `POST /api/v1/auth/reset-password` - Reset password with token

### Profile Management (Protected Routes)

- `GET /api/v1/profile` - Get current user profile
- `PUT /api/v1/profile` - Update current user profile
- `POST /api/v1/profile/change-password` - Change user password
- `POST /api/v1/profile/logout` - User logout

### User Management (Protected Routes)

- `GET /api/v1/users` - Get all users (paginated)
- `GET /api/v1/users/:id` - Get user by ID
- `POST /api/v1/users` - Create new user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Static Assets

- `GET /assets/*` - Serve static files from assets directory

## 🔧 Available Commands

### Database Commands

```bash
# Run migrations
go run cmd/migrate/main.go migrate

# Rollback migrations
go run cmd/migrate/main.go rollback

# Fresh migrations (rollback + migrate)
go run cmd/migrate/main.go fresh
```

### Development Commands

```bash
# Start server
go run cmd/server/main.go

# Build server
go build -o bin/server cmd/server/main.go

# Build migration tool
go build -o bin/migrate cmd/migrate/main.go

# Run tests
go test ./...

# Check for issues
go vet ./...
```

## 🏛️ Clean Architecture Benefits

### 1. **Testability**

- Easy to unit test business logic in isolation
- Mock interfaces for testing
- No external dependencies in core logic

### 2. **Maintainability**

- Clear separation of concerns
- Changes in one layer don't affect others
- Easy to understand and modify

### 3. **Flexibility**

- Easy to swap implementations (database, web framework, etc.)
- Can add new interfaces without changing core logic
- Framework-independent business logic

### 4. **Scalability**

- Well-organized code structure
- Easy to add new features
- Clear boundaries between components

## 🧪 Testing

The architecture makes testing straightforward:

```go
// Example: Testing use case with mocked dependencies
func TestUserUseCase_Login(t *testing.T) {
    // Arrange
    mockUserRepo := &mocks.UserRepository{}
    mockRefreshTokenRepo := &mocks.RefreshTokenRepository{}
    mockJWTService := &mocks.JWTService{}
    mockPasswordService := &mocks.PasswordService{}
    mockEmailUseCase := &mocks.EmailUseCase{}

    useCase := usecases.NewUserUseCase(
        mockUserRepo,
        mockRefreshTokenRepo,
        mockJWTService,
        mockPasswordService,
        mockEmailUseCase,
    )

    // Act & Assert
    // Test your business logic
}
```

## 🔐 Security Features

### Password Security

- **Bcrypt Hashing**: Industry-standard password hashing
- **Salt Generation**: Automatic salt generation for each password
- **Cost Factor**: Configurable cost factor for security vs performance

### JWT Security

- **HMAC Signing**: Secure token signing with secret keys
- **Token Expiration**: Configurable expiration times
- **Refresh Rotation**: Secure refresh token rotation

### Data Encryption

- **AES Encryption**: Additional data encryption capabilities
- **Configurable Keys**: Environment-based encryption keys
- **PKCS7 Padding**: Standard padding for block cipher

## 🔒 Authentication

The application uses JWT-based authentication with the following features:

- **Access Token**: Short-lived token for API requests (1 hour)
- **Refresh Token**: Long-lived token stored in HTTP-only cookie (7 days)
- **Password Hashing**: Bcrypt for secure password storage
- **AES Encryption**: Additional data encryption capabilities

### Authentication Flow

1. **Login**: User provides email/password, receives access token and refresh token (in cookie)
2. **API Requests**: Include access token in Authorization header
3. **Token Refresh**: Automatic refresh using HTTP-only cookie
4. **Logout**: Clears refresh token and invalidates session

Include the access token in requests:

```
Authorization: Bearer <access_token>
```

## 📧 Email Features

The application includes comprehensive email functionality:

- **Email Verification**: Send verification emails to new users
- **Password Reset**: Send password reset emails with secure tokens
- **Template System**: HTML email templates with dynamic data
- **SMTP Integration**: Configurable SMTP service for email delivery

## 📁 File Upload

Local file storage implementation:

- **Avatar Upload**: Users can upload profile pictures
- **Local Storage**: Files stored in local filesystem
- **File Validation**: Type and size validation
- **Secure Paths**: Protected file path handling

## 🛠️ Development Guidelines

### Adding New Features

1. **Start with Domain**: Define entities, value objects, and domain rules
2. **Define Ports**: Create interfaces for new repositories or services in `internal/core/ports/`
3. **Implement Use Cases**: Add business logic in `internal/core/usecases/`
4. **Create Adapters**: Implement interfaces in `internal/adapters/secondary/`
5. **Add HTTP Layer**: Create handlers in `internal/adapters/primary/http/handlers/`
6. **Wire Dependencies**: Update `internal/infrastructure/container.go`
7. **Update Routes**: Add new routes in `internal/adapters/primary/http/routes.go`

### Error Handling

- Domain errors defined in `internal/core/domain/errors/`
- Structured error responses with consistent format
- Proper HTTP status codes for different error types
- Error messages managed in `internal/adapters/primary/http/messages/`

### Service Implementations

**Current Services:**

- **UserUseCase**: User registration, login, profile management
- **EmailUseCase**: Email verification and password reset
- **JWTService**: Token generation and validation
- **PasswordService**: Bcrypt password hashing
- **EncryptionService**: AES encryption/decryption
- **MailerService**: SMTP email sending with templates
- **MediaService**: Local file storage and upload

## 📈 Performance & Production

### Build for Production

```bash
# Build optimized binary
go build -ldflags="-s -w" -o bin/server cmd/server/main.go

# Set production environment
export ENVIRONMENT=production
```

### Docker (Optional)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/templates ./templates
CMD ["./server"]
```

## 🌟 Features

### Core Features

- ✅ User Registration and Authentication
- ✅ JWT Token-based Authentication with Refresh Tokens
- ✅ Email Verification System
- ✅ Password Reset Functionality
- ✅ User Profile Management
- ✅ File Upload (Avatar)
- ✅ Pagination Support
- ✅ CORS Configuration
- ✅ Middleware Authentication

### Security Features

- ✅ Bcrypt Password Hashing
- ✅ JWT Token Security
- ✅ AES Data Encryption
- ✅ HTTP-Only Cookie for Refresh Tokens
- ✅ Input Validation and Sanitization

### Infrastructure Features

- ✅ Database Migration System
- ✅ Configuration Management
- ✅ SMTP Email Integration
- ✅ Local File Storage
- ✅ Structured Logging
- ✅ Graceful Shutdown
