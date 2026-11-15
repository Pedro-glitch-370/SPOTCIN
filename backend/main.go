package main

import (
	"log"
	"net/http"

	"SPOTCIN/backend/routes"
)

func main() {
	mux := routes.Setup()

	addr := ":8080"
	log.Printf("Servidor rodando em http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
