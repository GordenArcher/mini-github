package main

import (
	"github.com/GordenArcher/mini-github/internal/config"
	"github.com/GordenArcher/mini-github/internal/db"
	"github.com/GordenArcher/mini-github/internal/log"
	"github.com/GordenArcher/mini-github/internal/mail"
	"github.com/GordenArcher/mini-github/internal/middleware"
	"github.com/GordenArcher/mini-github/internal/redis"
	"github.com/GordenArcher/mini-github/internal/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	cfg := config.Load()

	log.Init()
	defer log.Sync()

	dbConn := db.Connect(cfg.DatabaseURL)
	dbConn.AutoMigrate(&db.User{}, &db.Repository{})

	redis.Connect(cfg.RedisAddr)

	mailer := mail.New(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)

	r := gin.Default()

	r.Use(middleware.RateLimitMiddleware())
	api := r.Group("/api/v1")

	// Auth API routes
	routes.RegisterAuthRoutes(api, dbConn, mailer, cfg)

	// Repo API routes
	middleware.SetJWTSecret(cfg.JWTAccessSecret)
	routes.RegisterRepoRoutes(api, dbConn)

	r.Run(":" + cfg.ServerPort)
}
