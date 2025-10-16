package config

import (
	"os"
)

type Config struct {
	DatabaseURL      string
	JWTAccessSecret  string
	JWTRefreshSecret string
	SMTPHost         string
	SMTPPort         string
	SMTPUser         string
	SMTPPass         string
	RedisAddr        string
	ServerPort       string
}

func Load() *Config {
	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "host=localhost user=postgres password=postgres dbname=mini_github port=5432 sslmode=disable TimeZone=UTC"),
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", "dev_access_secret"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "dev_refresh_secret"),
		SMTPHost:         getEnv("SMTP_HOST", ""),
		SMTPPort:         getEnv("SMTP_PORT", "587"),
		SMTPUser:         getEnv("SMTP_USER", ""),
		SMTPPass:         getEnv("SMTP_PASS", ""),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
