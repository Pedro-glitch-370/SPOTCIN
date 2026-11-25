package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"

	"SPOTCIN/backend/internal/services"
)

func Register(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" || len(body.Password) < 6 {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": "dados_invalidos"})
			return
		}
		user, err := services.RegisterUser(db, body.Email, body.Password)
		if err != nil {
			if errors.Is(err, services.ErrEmailTaken) {
				writeJSON(w, http.StatusConflict, map[string]string{"erro": "email_ja_cadastrado"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"erro": "internal_error"})
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"id": user.ID, "email": user.Email})
	}
}

func Login(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"erro": "dados_invalidos"})
			return
		}
		token, _, err := services.LoginUser(db, body.Email, body.Password)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"erro": "credenciais_invalidas"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"token": token})
	}
}