package database

import (
	"os"

	"SPOTCIN/backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// abre a conexão com o PostgreSQL e roda o AutoMigrate.
func Open() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=spotcin port=5432 sslmode=disable TimeZone=America/Sao_Paulo"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&models.Song{}); err != nil {
		return nil, err
	}

	return db, nil
}