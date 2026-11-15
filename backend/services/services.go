package services

import "SPOTCIN/backend/models"

// GetAllSongs lista as músicas disponíveis.
// MOCK: hardcoded.
func GetAllSongs() ([]models.Song, error) {
	return []models.Song{
		{
			ID:     "1",
			Title:  "Kokiri Forest",
			Artist: "Zelda",
		},
	}, nil
}

// GetSongPath retorna o caminho do arquivo de áudio pelo ID.
// MOCK: trocar no futuro
func GetSongPath(id string) (string, error) {
	switch id {
	case "1":
		return "./database/kokiri_forest.wav", nil
	default:
		return "", nil
	}
}