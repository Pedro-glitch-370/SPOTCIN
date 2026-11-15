package routes

import (
	"net/http"

	"SPOTCIN/backend/handlers"
	"SPOTCIN/backend/middleware"
)

// monta o mux, registra as rotas e envolve com os middlewares
func Setup() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("GET /api/songs", handlers.ListSongs)
	mux.HandleFunc("GET /api/stream/{id}", handlers.StreamSong)

	// aplica middlewares em cadeia
	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.Logging(h)

	return h
}