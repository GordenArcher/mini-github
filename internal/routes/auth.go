package routes

import (
	"time"

	"github.com/GordenArcher/mini-github/internal/config"
	"github.com/GordenArcher/mini-github/internal/db"
	"github.com/GordenArcher/mini-github/internal/handlers"
	"github.com/GordenArcher/mini-github/internal/mail"
	"github.com/GordenArcher/mini-github/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers auth-related endpoints
func RegisterAuthRoutes(r *gin.RouterGroup, dbConn *db.DB, mailer *mail.Mailer, cfg *config.Config) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", handlers.RegisterHandler(dbConn, mailer))
		auth.GET("/verify", handlers.VerifyEmailHandler(dbConn))
		auth.POST("/resend-verification", handlers.ResendVerificationHandler(dbConn, mailer))
		auth.POST("/login", handlers.LoginHandler(dbConn, cfg.JWTAccessSecret, cfg.JWTRefreshSecret, time.Hour*24*7))
		auth.POST("/refresh", handlers.RefreshHandler(dbConn, cfg.JWTAccessSecret, cfg.JWTRefreshSecret, time.Minute*15))
		auth.POST("/logout", handlers.LogoutHandler())
		auth.POST("/request-password-reset", handlers.RequestPasswordResetHandler(dbConn, mailer))
		auth.POST("/reset-password", handlers.ResetPasswordHandler(dbConn))
	}

	protected := r.Group("")
	protected.Use(middleware.JWTAuthMiddleware(cfg.JWTAccessSecret))
	{
		protected.GET("/me", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	}
}
