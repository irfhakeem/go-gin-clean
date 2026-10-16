# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binaries
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o main ./cmd/server
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o migrate ./cmd/migrate

# Production stage
FROM alpine:3.21

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binaries and migrations from builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY --from=builder /app/migrations ./migrations

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

EXPOSE 3000

# Run the binary
CMD ["./main"]
