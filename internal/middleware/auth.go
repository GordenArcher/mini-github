package middleware

import (
	"net/http"
	"strings"

	"github.com/GordenArcher/mini-github/internal/helper/responses"
	"github.com/GordenArcher/mini-github/internal/redis"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtAccessSecret string

// SetJWTSecret initializes the JWT secret
func SetJWTSecret(secret string) {
	jwtAccessSecret = secret
}

// AuthMiddleware is a shorthand wrapper for JWTAuthMiddleware using the stored secret
func AuthMiddleware() gin.HandlerFunc {
	if jwtAccessSecret == "" {
		panic("JWT access secret not set. Call middleware.SetJWTSecret(secret) first")
	}

	return JWTAuthMiddleware(jwtAccessSecret)
}

// JWTAuthMiddleware handles JWT validation
func JWTAuthMiddleware(accessSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			responses.JSONError(c, http.StatusUnauthorized, "authorization header required")
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			responses.JSONError(c, http.StatusUnauthorized, "authorization header format must be Bearer {token}")
			return
		}
		tokenString := parts[1]

		// parse token
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenMalformed
			}
			return []byte(accessSecret), nil
		})
		if err != nil || !token.Valid {
			responses.JSONError(c, http.StatusUnauthorized, "invalid token")
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			responses.JSONError(c, http.StatusUnauthorized, "invalid token claims")
			return
		}

		// check jti blacklist
		if jti, ok := claims["jti"].(string); ok {
			black, err := redis.IsAccessTokenBlacklisted(jti)
			if err == nil && black {
				responses.JSONError(c, http.StatusUnauthorized, "token revoked")
				return
			}
		}

		// extract sub
		sub, ok := claims["sub"].(float64)
		if !ok {
			responses.JSONError(c, http.StatusUnauthorized, "invalid token subject")
			return
		}
		c.Set("user_id", uint(sub))
		c.Next()
	}
}
