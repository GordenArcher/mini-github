package verifications

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/GordenArcher/mini-github/internal/db"
	"github.com/GordenArcher/mini-github/internal/log"
	"github.com/GordenArcher/mini-github/internal/mail"
	"github.com/GordenArcher/mini-github/internal/redis"
	"go.uber.org/zap"
)

func SendVerificationEmail(user *db.User, mailer *mail.Mailer) error {
	tok := make([]byte, 32)
	if _, err := rand.Read(tok); err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	verif := hex.EncodeToString(tok)

	// store in redis with TTL
	if err := redis.Client.Set(redis.Ctx, "verify:"+verif, user.ID, time.Hour*24).Err(); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}

	link := fmt.Sprintf("http://localhost:8080/api/v1/auth/verify?token=%s", verif)
	body := fmt.Sprintf("Please verify your email by clicking <a href=\"%s\">here</a>", link)

	if mailer != nil {
		go func() {
			if err := mailer.Send(user.Email, "Verify your account", body); err != nil {
				log.Logger.Error("failed to send verification email", zap.Error(err))
			} else {
				log.Logger.Info("verification email sent", zap.String("to", user.Email))
			}
		}()
	}

	return nil
}
