package middleware

import (
	"strings"

	"go-gin-clean/internal/gateway/security"
	pkgerror "go-gin-clean/pkg/error"
	"go-gin-clean/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtService security.JWTServiceInterface
}

func NewAuthMiddleware(jwtService security.JWTServiceInterface) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, pkgerror.Unauthorized(pkgerror.ErrAuthHeaderMissing, nil))
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			response.Error(c, pkgerror.Unauthorized(pkgerror.ErrAuthHeaderMissing, nil))
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			response.Error(c, pkgerror.Unauthorized(pkgerror.ErrTokenNotFound, nil))
			c.Abort()
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			response.Error(c, pkgerror.Unauthorized(pkgerror.ErrTokenInvalid, err))
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID.String())
		c.Set("user_role", claims.UserRole)

		c.Next()
	}
}
