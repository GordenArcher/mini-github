package db

import (
	"fmt"
	"time"

	"github.com/GordenArcher/mini-github/internal/log"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"uniqueIndex;not null" json:"username"`
	Email       string    `gorm:"uniqueIndex;not null" json:"email"`
	Password    string    `gorm:"not null" json:"-"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Bio         string    `json:"bio"`
	IsVerified  bool      `gorm:"default:false" json:"is_verified"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Repository struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description string
	Visibility  string `gorm:"default:'private'"` // "private" or "public"
	Path        string `gorm:"not null"`          // local path on server
	OwnerID     uint
	Owner       User `gorm:"foreignKey:OwnerID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (u *User) AfterCreate(tx *gorm.DB) (err error) {
	defaultRepo := Repository{
		Name:       fmt.Sprintf("%s-first-repo", u.Username),
		OwnerID:    u.ID,
		Visibility: "private",
	}

	if err := tx.Create(&defaultRepo).Error; err != nil {
		log.Logger.Error("failed to create default repo for user", zap.Error(err))
		return err
	}

	log.Logger.Info("default repo created", zap.String("user", u.Username))
	return nil
}
