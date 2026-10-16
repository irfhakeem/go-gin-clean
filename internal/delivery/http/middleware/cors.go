package middleware

import (
	"slices"

	"github.com/gin-gonic/gin"
)

func CORS(allowedOrigins ...string) gin.HandlerFunc {
	origins := allowedOrigins
	if len(origins) == 0 {
		origins = []string{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if origin != "" {
			isAllowed := false
			if slices.Contains(origins, "*") {
				c.Header("Access-Control-Allow-Origin", "*")
				isAllowed = true
			} else if slices.Contains(origins, origin) {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				isAllowed = true
			}

			if isAllowed {
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Origin, Content-Length, Content-Type, Authorization, X-Refresh-Token, X-Request-ID")
				c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
			}
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
