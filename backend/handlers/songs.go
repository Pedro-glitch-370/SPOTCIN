package handlers

import (
	"net/http"
	"os"

	"SPOTCIN/backend/services"
)

// responde com a lista de músicas em JSON.
func ListSongs(w http.ResponseWriter, r *http.Request) {
	songs, err := services.GetAllSongs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// serialize songs em JSON...
	w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(songs)
	_ = songs
}

// serve o áudio pelo ID.
func StreamSong(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	path, err := services.GetSongPath(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	f, err := os.Open(path)
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

	// trata Range Requests automaticamente → 206 Partial Content
	http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
}
