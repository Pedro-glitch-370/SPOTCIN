package database

import (
	"errors"
	"os"

	"SPOTCIN/backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// abre a conexão com o PostgreSQL (se tiver setado. se não, abre SQLite) e roda o AutoMigrate.
func Open() (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)

	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("DATABASE_URL não definida")
	}

	if err := db.AutoMigrate(&models.Song{}, &models.User{}); err != nil {
		return nil, err
	}

	return db, nil
}
