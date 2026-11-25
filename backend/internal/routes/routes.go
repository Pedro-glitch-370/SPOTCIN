package routes

import (
	"net/http"

	"SPOTCIN/backend/internal/handlers"
	"SPOTCIN/backend/internal/middleware"

	"gorm.io/gorm"
)

// Setup monta o mux, registra as rotas e envolve com os middlewares
// A ordem é CORS, RequestID, Logging e então mux
func Setup(db *gorm.DB) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("GET /health/ready", handlers.Ready(db))
	mux.HandleFunc("POST /api/auth/register", handlers.Register(db))
	mux.HandleFunc("POST /api/auth/login", handlers.Login(db))
	mux.HandleFunc("GET /api/songs", handlers.ListSongs(db))
	mux.HandleFunc("GET /api/songs/{id}", handlers.GetSong(db))
	mux.HandleFunc("GET /api/stream/{id}", handlers.StreamSong(db))
	mux.Handle("GET /api/playback/{id}", middleware.Auth(http.HandlerFunc(handlers.GetPlayback(db))))
	mux.Handle("PUT /api/playback/{id}", middleware.Auth(http.HandlerFunc(handlers.PutPlayback(db))))

	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.RequestID(h)
	h = middleware.CORS(h)

	return h
}