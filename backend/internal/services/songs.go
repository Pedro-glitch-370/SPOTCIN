package services

import (
	"errors"

	"SPOTCIN/backend/internal/models"

	"gorm.io/gorm"
)

// ErrSongNotFound é retornado quando o ID não existe no DB
var ErrSongNotFound = errors.New("música não encontrada")

// GetAllSongs lista todas as músicas indexadas.
func GetAllSongs(db *gorm.DB) ([]models.Song, error) {
	var songs []models.Song
	if err := db.Find(&songs).Error; err != nil {
		return nil, err
	}
	return songs, nil
}

// GetSongPath retorna o caminho absoluto e o MIME type do ID informado
// O path no DB é relativo; aqui montamos o absoluto com AudioDir
func GetSongPath(db *gorm.DB, id string) (absPath, mime string, err error) {
	var song models.Song
	if err := db.First(&song, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", ErrSongNotFound
		}
		return "", "", err
	}
	return song.Path, song.MimeType, nil
}