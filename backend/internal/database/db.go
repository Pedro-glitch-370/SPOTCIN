package database

import (
	"os"

	"SPOTCIN/backend/internal/models"

	"github.com/glebarez/sqlite"
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
		// esse SQLite é apenas temporário, o arquivo fica no diretório de trabalho atual
		db, err = gorm.Open(sqlite.Open("spotcin.db"), &gorm.Config{})
		if err != nil {
			return nil, err
		}
	}

	if err := db.AutoMigrate(&models.Song{}); err != nil {
		return nil, err
	}

	return db, nil
}
