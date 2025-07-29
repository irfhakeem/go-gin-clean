package http

import (
	"go-gin-clean/internal/core/ports"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	router *gin.Engine,
	userUseCase ports.UserUseCase,
	jwtService ports.JWTService,
) {
	// Setup handlers
	userHandler := NewUserHandler(userUseCase)
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
			auth.GET("/verify-email", userHandler.VerifyEmail)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			protected.POST("/logout", userHandler.Logout)
			protected.POST("/change-password", userHandler.ChangePassword)

			// User management routes (admin/protected)
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

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"message": "Server is running",
		})
	})
}
