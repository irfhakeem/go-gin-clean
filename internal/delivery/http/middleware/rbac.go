package middleware

import (
	"go-gin-clean/internal/delivery/http/response"
	"go-gin-clean/internal/domain/permission"
	"go-gin-clean/internal/domain/policy"
	pkgerrors "go-gin-clean/pkg/errors"
	"go-gin-clean/pkg/message"

	"github.com/gin-gonic/gin"
)

func RequireRole(checker permission.Checker, required permission.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("actor")
		if !exists {
			response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrAuthHeaderMissing))
			c.Abort()
			return
		}

		actor, ok := val.(policy.Actor)
		if !ok || !checker.HasPermission(actor, required) {
			response.Error(c, pkgerrors.NewAppError(pkgerrors.Forbidden, message.ErrForbidden))
			c.Abort()
			return
		}

		c.Next()
	}
}
