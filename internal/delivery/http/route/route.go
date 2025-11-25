package route

import (
	"go-gin-clean/internal/delivery/http"
	"go-gin-clean/internal/delivery/http/middleware"
	"go-gin-clean/internal/gateway/security"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	router *gin.Engine,
	userHandler *http.UserHandler,
	oauthHandler *http.OAuthHandler,
	jwtService *security.JWTService,
) {
	// Setup handlers
	authMiddleware := middleware.NewAuthMiddleware(jwtService)

	// Setup CORS
	router.Use(middleware.CORS())

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes (auth)
		auth := api.Group("/auth")
		{
			auth.POST("/login", userHandler.Login)
			auth.POST("/register", userHandler.Register)
			auth.POST("/refresh-token", userHandler.RefreshToken)
			auth.POST("/verify-email", userHandler.VerifyEmail)
			auth.POST("/reset-password", userHandler.ResetPassword)
			auth.POST("/send-reset-password", userHandler.SendResetPassword)
			auth.POST("/resend-verification", userHandler.SendVerifyEmail)
		}

		oauth := auth.Group("/oauth2")
		{
			oauth.POST("/url", oauthHandler.GetLoginURL)
			oauth.GET("/:provider/callback", oauthHandler.CallBack)
		}

		profile := api.Group("/profile")
		profile.Use(authMiddleware.RequireAuth())
		{
			profile.GET("", userHandler.Profile)
			profile.PUT("", userHandler.UpdateProfile)
			profile.PUT("/change-password", userHandler.ChangePassword)
			profile.POST("/logout", userHandler.Logout)
		}

		users := api.Group("/users")
		users.Use(authMiddleware.RequireAuth())
		{
			users.GET("", userHandler.GetAllUsers)
			users.GET("/:code", userHandler.GetUserByCode)
			users.POST("", userHandler.CreateUser)
			users.PUT("/:code", userHandler.UpdateUser)
			users.PUT("/:code/change-status", userHandler.ChangeStatus)
			users.DELETE("/:code", userHandler.DeleteUser)
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"message": "Server is running",
		})
	})
}
