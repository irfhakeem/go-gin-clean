package middleware

import (
	"strings"

	"go-gin-clean/internal/application/port"
	"go-gin-clean/internal/delivery/http/response"
	pkgerrors "go-gin-clean/pkg/errors"
	"go-gin-clean/pkg/message"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	token port.TokenMaker
}

func NewAuthMiddleware(token port.TokenMaker) *AuthMiddleware {
	return &AuthMiddleware{
		token: token,
	}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrAuthHeaderMissing))
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrAuthHeaderMissing))
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrTokenNotFound))
			c.Abort()
			return
		}

		claims, err := m.token.ValidateAccessToken(token)
		if err != nil {
			response.Error(c, pkgerrors.AsAppError(pkgerrors.Unauthorized, message.ErrTokenInvalid, err))
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID.String())
		c.Set("user_role", claims.UserRole)

		c.Next()
	}
}
