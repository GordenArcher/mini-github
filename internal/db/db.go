package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct{ *gorm.DB }

func Connect(dsn string) *DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	return &DB{DB: db}
}
