package main

import (
	"fmt"
	"net/http"
)

// HTTP básico para testar o servidor
func main() {
    http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, `{"status":"ok"}`)
    })
    http.ListenAndServe(":8080", nil)
}