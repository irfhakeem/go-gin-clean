package http

import (
	"go-gin-clean/internal/adapters/primary/http/messages"
	"go-gin-clean/internal/adapters/primary/http/response"
	"go-gin-clean/internal/core/domain/errors"
	"go-gin-clean/internal/core/ports"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtService ports.JWTService
}

func NewAuthMiddleware(jwtService ports.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, messages.FAILED_AUTHENTICATION_REQUIRED, errors.ErrAuthHeaderMissing.Error(), http.StatusUnauthorized)
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			response.Error(c, messages.FAILED_INVALID_TOKEN_FORMAT, errors.ErrAuthHeaderMissing.Error(), http.StatusUnauthorized)
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			response.Error(c, messages.FAILED_TOKEN_NOT_FOUND, errors.ErrTokenNotFound.Error(), http.StatusUnauthorized)
			c.Abort()
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			response.Error(c, messages.FAILED_INVALID_TOKEN_FORMAT, errors.ErrTokenInvalid.Error(), http.StatusUnauthorized)
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Authorization, X-Refresh-Token")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
