package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"gorm.io/gorm"

	"SPOTCIN/backend/internal/database"
)

type healthStatus struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

// Health confirma que o processo tá vivo
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Write([]byte(`{"status":"ok"}`))
}

// Ready checa Postgres, Redis e Garage
func Ready(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		checks := map[string]string{}
		ok := true

		// Postgre
		if sqlDB, err := db.DB(); err != nil {
			checks["postgres"] = "fail: " + err.Error()
			ok = false
		} else if err := sqlDB.PingContext(ctx); err != nil {
			checks["postgres"] = "fail: " + err.Error()
			ok = false
		} else {
			checks["postgres"] = "ok"
		}

		// Redis
		if database.RDB == nil {
			checks["redis"] = "fail: cliente não inicializado"
			ok = false
		} else if err := database.RDB.Ping(ctx).Err(); err != nil {
			checks["redis"] = "fail: " + err.Error()
			ok = false
		} else {
			checks["redis"] = "ok"
		}

		// Garage
		if database.GarageClient == nil {
			checks["garage"] = "fail: cliente não inicializado"
			ok = false
		} else {
			bucket := "spotcin-audio"
			if _, err := database.GarageClient.BucketExists(ctx, bucket); err != nil {
				checks["garage"] = "fail: " + err.Error()
				ok = false
			} else {
				checks["garage"] = "ok"
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if !ok {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(healthStatus{
			Status: map[bool]string{true: "ok", false: "degraded"}[ok],
			Checks: checks,
		})
	}
}