package main

import (
	"SPOTCIN/backend/internal/database"
	"SPOTCIN/backend/internal/routes"
	"SPOTCIN/backend/internal/services"
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))
	
	db, err := database.Open()
	if err != nil {
		log.Fatalf("abrindo DB: %v", err)
	}

	if err := database.OpenRedis(); err != nil {
		log.Fatalf("abrindo Redis: %v", err)
	}

	if err := database.OpenGarage(); err != nil {
		log.Fatalf("abrindo Garage: %v", err)
	}
	
	if err := services.SeedGarage(context.Background()); err != nil {
		log.Printf("seed garage falhou: %v", err)
	}

	if err := services.ScanDatabase(db); err != nil {
		log.Printf("scan inicial falhou: %v", err)
	}

	addr := ":8080"
	log.Printf("Servidor rodando em http://localhost%s", addr)
	if err := http.ListenAndServe(addr, routes.Setup(db)); err != nil {
		log.Fatal(err)
	}
}