package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"

	"SPOTCIN/backend/internal/services"
)

// GetPlayback responde com a posição salva pro (user_id, track_id)
func GetPlayback(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "user_id_ausente"})
			return
		}

		id := r.PathValue("id")
		p, err := services.GetPlayback(r.Context(), id, userID)
		if err != nil {
			if errors.Is(err, services.ErrPlaybackNotFound) {
				writeJSON(w, http.StatusNotFound, map[string]string{"erro": "posicao_nao_encontrada"})
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, p)
	}
}

// PutPlayback valida a faixa no postgre, valida o body e grava no Redis
func PutPlayback(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"erro": "user_id_ausente"})
			return
		}

		id := r.PathValue("id")

		// valida a faixa no postgre
		if _, _, err := services.GetSongPath(db, id); err != nil {
			if errors.Is(err, services.ErrSongNotFound) {
				writeJSON(w, http.StatusNotFound, map[string]string{"erro": "track_nao_encontrada"})
				return
			}
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		var body struct {
			PosicaoSeg *float64 `json:"posicao_seg"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.PosicaoSeg == nil || *body.PosicaoSeg < 0 {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": "posicao_invalida"})
			return
		}

		if err := services.SavePlayback(r.Context(), id, userID, *body.PosicaoSeg); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// writeJSON escreve JSON com status
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}