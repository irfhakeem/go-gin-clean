package route

import (
	"github.com/gin-gonic/gin"

	"go-gin-clean/internal/application/port"
	"go-gin-clean/internal/delivery/http"
	"go-gin-clean/internal/delivery/http/middleware"
	"go-gin-clean/internal/domain/permission"
)

func SetupRoutes(
	router *gin.Engine,
	token port.TokenMaker,
	checker permission.Checker,
	auth *http.AuthHandler,
	user *http.UserHandler,
) {
	authMiddleware := middleware.NewAuthMiddleware(token)

	router.Use(middleware.CORS())

	api := router.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", auth.Login)
			authGroup.POST("/register", auth.Register)
			authGroup.POST("/refresh-token", auth.RefreshToken)
			authGroup.POST("/verify-email", auth.VerifyEmail)
			authGroup.POST("/reset-password", auth.ResetPassword)
			authGroup.POST("/send-reset-password", auth.SendResetPassword)
			authGroup.POST("/resend-verification", auth.SendVerifyEmail)
		}

		oauth2 := authGroup.Group("/oauth2")
		{
			oauth2.POST("/url", auth.GetOAuthLoginURL)
			oauth2.GET("/:provider/callback", auth.HandleOAuthCallback)
		}

		profile := api.Group("/profile")
		profile.Use(authMiddleware.RequireAuth())
		{
			profile.GET("", user.Profile)
			profile.PUT("", user.UpdateProfile)
			profile.PUT("/change-password", user.ChangePassword)
			profile.POST("/logout", auth.Logout)
		}

		users := api.Group("/users")
		users.Use(authMiddleware.RequireAuth())
		{
			users.GET("", middleware.RequireRole(checker, permission.UserRead), user.GetAllUsers)
			users.GET("/:id", middleware.RequireRole(checker, permission.UserRead), user.GetUserByID)
			users.POST("", middleware.RequireRole(checker, permission.UserCreate), user.CreateUser)
			users.PUT("/:id", middleware.RequireRole(checker, permission.UserUpdate), user.UpdateUser)
			users.PUT("/:id/change-status", middleware.RequireRole(checker, permission.UserUpdate), user.ChangeStatus)
			users.DELETE("/:id", middleware.RequireRole(checker, permission.UserDelete), user.DeleteUser)
		}
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"message": "Server is running",
		})
	})
}
