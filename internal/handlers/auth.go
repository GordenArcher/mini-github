package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/GordenArcher/mini-github/internal/db"
	"github.com/GordenArcher/mini-github/internal/errors"
	"github.com/GordenArcher/mini-github/internal/helper/responses"
	"github.com/GordenArcher/mini-github/internal/helper/utils/verifications"
	"github.com/GordenArcher/mini-github/internal/log"
	"github.com/GordenArcher/mini-github/internal/mail"
	"github.com/GordenArcher/mini-github/internal/redis"
)

func RegisterHandler(database *db.DB, mailer *mail.Mailer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload struct {
			Username string `json:"username" binding:"required,min=3"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
		}

		if err := c.ShouldBindJSON(&payload); err != nil {
			errors.AbortWithError(c, http.StatusBadRequest, err.Error())
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			errors.AbortWithError(c, 500, "failed to hash password")
			return
		}

		var existing db.User
		if err := database.Where("email = ?", payload.Email).First(&existing).Error; err == nil {
			responses.JSONError(c, http.StatusBadRequest, "email already registered")
			return
		}

		var existing2 db.User
		if err := database.Where("username = ?", payload.Username).First(&existing2).Error; err == nil {
			responses.JSONError(c, http.StatusBadRequest, "username already in use")
			return
		}

		user := db.User{Username: payload.Username, Email: payload.Email, Password: string(hash)}
		if err := database.Create(&user).Error; err != nil {
			responses.JSONError(c, http.StatusBadRequest, "error creating user")
			return
		}

		if err := verifications.SendVerificationEmail(&user, mailer); err != nil {
			log.Logger.Error("failed to send verification", zap.Error(err))
		}

		log.Logger.Info("user registered", zap.String("email", user.Email))
		responses.JSONSuccess(c, http.StatusCreated, "registered successful", err)
	}
}

func VerifyEmailHandler(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		t := c.Query("token")

		if t == "" {
			errors.AbortWithError(c, 400, "token required")
			return
		}
		val, err := redis.Client.Get(redis.Ctx, "verify:"+t).Result()

		if err != nil {
			errors.AbortWithError(c, 400, "invalid or expired token")
			return
		}
		var id uint
		fmt.Sscanf(val, "%d", &id)

		if err := database.Model(&db.User{}).Where("id = ?", id).Update("is_verified", true).Error; err != nil {
			errors.AbortWithError(c, 500, "failed to verify")
			return
		}

		redis.Client.Del(redis.Ctx, "verify:"+t)
		c.JSON(http.StatusOK, gin.H{"verified": true})
	}
}

func ResendVerificationHandler(database *db.DB, mailer *mail.Mailer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload struct {
			Email string `json:"email"`
		}

		if err := c.ShouldBindJSON(&payload); err != nil {
			responses.JSONError(c, http.StatusBadRequest, "invalid request payload")
			return
		}

		var user db.User
		if err := database.Where("email = ?", payload.Email).First(&user).Error; err != nil {
			responses.JSONError(c, http.StatusNotFound, "user not found")
			return
		}

		if user.IsVerified {
			responses.JSONError(c, http.StatusBadRequest, "account already verified")
			return
		}

		if err := verifications.SendVerificationEmail(&user, mailer); err != nil {
			responses.JSONError(c, http.StatusInternalServerError, "failed to send verification email")
			return
		}

		responses.JSONSuccess(c, http.StatusOK, "verification email resent", nil)
	}
}

func LoginHandler(database *db.DB, accessSecret, refreshSecret string, refreshTTL time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload struct{ Email, Password string }
		if err := c.ShouldBindJSON(&payload); err != nil {
			errors.AbortWithError(c, 400, err.Error())
			return
		}

		var user db.User
		if err := database.Where("email = ?", payload.Email).First(&user).Error; err != nil {
			responses.JSONError(c, 401, "invalid credentials")
			return
		}

		if !user.IsVerified {
			responses.JSONError(c, http.StatusForbidden, "please verify your email before logging in")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
			responses.JSONError(c, 401, "invalid credentials")
			return
		}

		// create access token with jti
		jtiBytes := make([]byte, 16)
		rand.Read(jtiBytes)
		jti := hex.EncodeToString(jtiBytes)
		access := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": user.ID, "exp": time.Now().Add(15 * time.Minute).Unix(), "jti": jti})
		accessString, _ := access.SignedString([]byte(accessSecret))

		// create refresh token (random) and store in redis
		refreshBytes := make([]byte, 64)
		rand.Read(refreshBytes)
		refreshToken := hex.EncodeToString(refreshBytes)
		redis.SetRefreshToken(refreshToken, user.ID, refreshTTL)

		responses.JSONSuccess(c, http.StatusOK, "login successful", gin.H{"access_token": accessString, "refresh_token": refreshToken, "token_type": "bearer", "expires_in": 900})
	}
}

func RefreshHandler(database *db.DB, accessSecret, refreshSecret string, accessTTL time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload struct {
			Refresh string `json:"refresh_token"`
		}

		if err := c.ShouldBindJSON(&payload); err != nil {
			errors.AbortWithError(c, 400, "refresh_token required")
			return
		}

		// validate refresh from redis
		val, err := redis.Client.Get(redis.Ctx, "refresh:"+payload.Refresh).Result()
		if err != nil {
			responses.JSONError(c, 401, "invalid refresh token")
			return
		}

		var id uint
		fmt.Sscanf(val, "%d", &id)

		// issue new access token
		jtiBytes := make([]byte, 16)
		rand.Read(jtiBytes)
		jti := hex.EncodeToString(jtiBytes)
		access := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": id, "exp": time.Now().Add(accessTTL).Unix(), "jti": jti})
		accessString, _ := access.SignedString([]byte(accessSecret))

		responses.JSONSuccess(c, http.StatusOK, "refreshed", gin.H{"access_token": accessString, "expires_in": int(accessTTL.Seconds())})
	}
}

func LogoutHandler() gin.HandlerFunc {
	// revokes refresh and blacklists access jti

	return func(c *gin.Context) {
		var payload struct {
			Refresh string `json:"refresh_token"`
		}

		if err := c.ShouldBindJSON(&payload); err != nil {
			errors.AbortWithError(c, 400, "refresh_token required")
			return
		}

		// revoke refresh
		redis.RevokeRefreshToken(payload.Refresh)

		// blacklist access jti from token in header
		auth := c.GetHeader("Authorization")
		if auth != "" {
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) == 2 {
				tokenString := parts[1]
				token, _ := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) { return []byte(""), nil })
				if tokClaims, ok := token.Claims.(jwt.MapClaims); ok {
					if jti, ok := tokClaims["jti"].(string); ok {
						redis.BlacklistAccessToken(jti, time.Hour*24)
					}
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{"logged_out": true})
	}
}

func RequestPasswordResetHandler(database *db.DB, mailer *mail.Mailer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload struct {
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			errors.AbortWithError(c, 400, "email required")
			return
		}

		var user db.User
		if err := database.Where("email = ?", payload.Email).First(&user).Error; err != nil {
			c.JSON(200, gin.H{"ok": true})
			return
		}

		// generate token
		tok := make([]byte, 32)
		rand.Read(tok)
		code := hex.EncodeToString(tok)
		redis.Client.Set(redis.Ctx, "pwreset:"+code, user.ID, time.Hour*1)
		link := fmt.Sprintf("http://localhost:8080/api/v1/auth/reset?token=%s", code)

		if mailer != nil {
			go func() {
				if err := mailer.Send(user.Email, "Password reset", fmt.Sprintf("Click <a href=\"%s\">here</a>", link)); err != nil {
					log.Logger.Error("failed to send password reset email", zap.Error(err))
				} else {
					log.Logger.Info("password reset email sent", zap.String("to", user.Email))
				}
			}()
		}

		responses.JSONSuccess(c, 200, "please chack your email", gin.H{"ok": true})
	}
}

func ResetPasswordHandler(database *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload struct{ Token, Password string }
		if err := c.ShouldBindJSON(&payload); err != nil {
			errors.AbortWithError(c, 400, "token and password required")
			return
		}
		val, err := redis.Client.Get(redis.Ctx, "pwreset:"+payload.Token).Result()
		if err != nil {
			errors.AbortWithError(c, 400, "invalid or expired token")
			return
		}

		var id uint
		fmt.Sscanf(val, "%d", &id)
		hash, _ := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		database.Model(&db.User{}).Where("id = ?", id).Update("password", string(hash))
		redis.Client.Del(redis.Ctx, "pwreset:"+payload.Token)

		c.JSON(200, gin.H{"ok": true})
	}
}
