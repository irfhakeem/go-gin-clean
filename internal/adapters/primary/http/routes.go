package http

import (
	"go-gin-clean/internal/adapters/primary/http/handlers"
	"go-gin-clean/internal/adapters/primary/http/mappers"
	"go-gin-clean/internal/core/ports"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	router *gin.Engine,
	userUseCase ports.UserUseCase,
	jwtService ports.JWTService,
) {
	// Setup mappers
	userMapper := mappers.NewUserMapper()

	// Setup handlers
	userHandler := handlers.NewUserHandler(userUseCase, userMapper)
	authMiddleware := NewAuthMiddleware(jwtService)

	// Setup CORS
	router.Use(CORS())

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
			auth.POST("/send-verify-email", userHandler.SendVerifyEmail)
			auth.POST("/reset-password", userHandler.ResetPassword)
			auth.POST("/send-reset-password", userHandler.SendResetPassword)
		}

		// Protected routes

		protected := api.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			// Profile routes
			profile := protected.Group("/profile")
			{
				profile.GET("", userHandler.Profile)
				profile.PUT("", userHandler.UpdateProfile)
				profile.POST("/change-password", userHandler.ChangePassword)
				profile.POST("/logout", userHandler.Logout)
			}

			// User management routes (protected)
			users := protected.Group("/users")
			{
				users.GET("", userHandler.GetAllUsers)
				users.GET("/:id", userHandler.GetUserByID)
				users.POST("", userHandler.CreateUser)
				users.PUT("/:id", userHandler.UpdateUser)
				users.DELETE("/:id", userHandler.DeleteUser)
			}
		}
	}

	router.Static("/assets", "./assets")

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  true,
			"message": "Server is running",
		})
	})
}
