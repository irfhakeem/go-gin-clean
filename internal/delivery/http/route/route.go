package route

import (
	"go-gin-clean/internal/application/port"
	"go-gin-clean/internal/delivery/http"
	"go-gin-clean/internal/delivery/http/middleware"
	"go-gin-clean/internal/domain/permission"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	router *gin.Engine,
	token port.TokenMaker,
	checker permission.Checker,
	user *http.UserHandler,
	oauth *http.OAuthHandler,
) {
	authMiddleware := middleware.NewAuthMiddleware(token)

	router.Use(middleware.CORS())

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", user.Login)
			auth.POST("/register", user.Register)
			auth.POST("/refresh-token", user.RefreshToken)
			auth.POST("/verify-email", user.VerifyEmail)
			auth.POST("/reset-password", user.ResetPassword)
			auth.POST("/send-reset-password", user.SendResetPassword)
			auth.POST("/resend-verification", user.SendVerifyEmail)
		}

		oauth2 := auth.Group("/oauth2")
		{
			oauth2.POST("/url", oauth.GetLoginURL)
			oauth2.GET("/:provider/callback", oauth.CallBack)
		}

		profile := api.Group("/profile")
		profile.Use(authMiddleware.RequireAuth())
		{
			profile.GET("", user.Profile)
			profile.PUT("", user.UpdateProfile)
			profile.PUT("/change-password", user.ChangePassword)
			profile.POST("/logout", user.Logout)
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
