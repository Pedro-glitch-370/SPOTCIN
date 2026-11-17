package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"SPOTCIN/backend/services"

	"gorm.io/gorm"
)

// ListSongs responde com a lista de músicas em JSON.
func ListSongs(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		songs, err := services.GetAllSongs(db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(songs); err != nil {
			http.Error(w, "erro ao serializar", http.StatusInternalServerError)
		}
	}
}

// StreamSong serve o áudio pelo ID. ServeContent cuida de Range (206).
func StreamSong(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		path, mime, err := services.GetSongPath(db, id)
		if err != nil {
			if errors.Is(err, services.ErrSongNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// valida path traversal contra AudioDir
		absClean, err := filepath.Abs(path)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		audioDirAbs, _ := filepath.Abs(services.AudioDir)
		if !strings.HasPrefix(absClean, audioDirAbs+string(os.PathSeparator)) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		f, err := os.Open(absClean)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", mime)
		http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
	}
}