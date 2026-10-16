package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type requestIDKeyType struct{}

var RequestIDContextKey = requestIDKeyType{}

const RequestIDHeader = "X-Request-ID"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set("request_id", requestID)
		c.Writer.Header().Set(RequestIDHeader, requestID)

		ctx := context.WithValue(c.Request.Context(), RequestIDContextKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func GetRequestID(c *gin.Context) string {
	if val, ok := c.Get("request_id"); ok {
		if id, ok := val.(string); ok {
			return id
		}
	}
	return ""
}
