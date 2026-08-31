package middleware

import (
	"go-gin-clean/internal/delivery/http/response"
	"go-gin-clean/internal/domain/permission"
	"go-gin-clean/internal/domain/policy"
	pkgerrors "go-gin-clean/pkg/errors"
	"go-gin-clean/pkg/message"

	"github.com/gin-gonic/gin"
)

func RequireRole(checker permission.Checker,
	required permission.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrAuthHeaderMissing))
			c.Abort()
			return
		}

		userRole, exists := c.Get("user_role")
		if !exists {
			response.Error(c, pkgerrors.NewAppError(pkgerrors.Unauthorized, message.ErrAuthHeaderMissing))
			c.Abort()
			return
		}

		actor, err := policy.NewActor(userID.(string), userRole.(string))
		if err != nil {
			response.Error(c, pkgerrors.WrapAppError(pkgerrors.Internal, message.ErrInvalidClaims, err))
			c.Abort()
			return
		}

		if !checker.HasPermission(actor, required) {
			response.Error(c, pkgerrors.NewAppError(pkgerrors.Forbidden, message.ErrForbidden))
			c.Abort()
			return
		}

		c.Next()
	}
}
