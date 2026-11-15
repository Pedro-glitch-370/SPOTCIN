package routes

import (
	"net/http"

	"SPOTCIN/backend/handlers"
	"SPOTCIN/backend/middleware"
)

func Setup() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health)

	// aplica middlewares em cadeia
	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.Logging(h)

	return h
}