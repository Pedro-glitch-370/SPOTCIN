package main

import (
	"log"
	"net/http"

	"SPOTCIN/backend/internal/database"
	"SPOTCIN/backend/internal/routes"
	"SPOTCIN/backend/internal/services"
)

func main() {
	db, err := database.Open()
	if err != nil {
		log.Fatalf("abrindo DB: %v", err)
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