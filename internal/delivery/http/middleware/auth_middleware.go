package middleware

import (
	"go-gin-clean/internal/delivery/http/response"
	"go-gin-clean/internal/gateway/security"
	"go-gin-clean/pkg/errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtService *security.JWTService
}

func NewAuthMiddleware(jwtService *security.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, "authentication required", errors.ErrAuthHeaderMissing.Error(), http.StatusUnauthorized)
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			response.Error(c, "invalid token format", errors.ErrAuthHeaderMissing.Error(), http.StatusUnauthorized)
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			response.Error(c, "token not found", errors.ErrTokenNotFound.Error(), http.StatusUnauthorized)
			c.Abort()
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			response.Error(c, "invalid token", errors.ErrTokenInvalid.Error(), http.StatusUnauthorized)
			c.Abort()
			return
		}

		c.Set("user_pkid", claims.UserPKID)
		c.Set("user_code", claims.UserCode)
		c.Set("user_role", claims.UserRole)

		c.Next()
	}
}
