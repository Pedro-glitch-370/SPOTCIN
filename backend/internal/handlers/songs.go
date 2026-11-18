package handlers

import (
	"SPOTCIN/backend/internal/database"
	"SPOTCIN/backend/internal/services"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
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

// StreamSong serve o áudio pelo ID, lendo do Garage.
// ServeContent cuida de Range (206 e 416).
func StreamSong(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		key, mime, err := services.GetSongPath(db, id)
		if err != nil {
			if errors.Is(err, services.ErrSongNotFound) {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// rejeita chaves com ".."
		if strings.Contains(key, "..") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		obj, err := database.GarageClient.GetObject(ctx, database.GarageBucket, key, minio.GetObjectOptions{})
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer obj.Close()

		info, err := obj.Stat()
		if err != nil {
			resp := minio.ToErrorResponse(err)
			if resp.Code == "NoSuchKey" {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", mime)
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		http.ServeContent(w, r, key, info.LastModified, obj)
	}
}