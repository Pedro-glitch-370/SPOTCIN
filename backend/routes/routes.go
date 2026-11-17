package routes

import (
	"net/http"

	"SPOTCIN/backend/handlers"
	"SPOTCIN/backend/middleware"

	"gorm.io/gorm"
)

// monta o mux, registra as rotas e envolve com os middlewares
func Setup(db *gorm.DB) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("GET /api/songs", handlers.ListSongs(db))
	mux.HandleFunc("GET /api/stream/{id}", handlers.StreamSong(db))

	// aplica middlewares em cadeia
	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.Logging(h)

	return h
}