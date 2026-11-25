package middleware

import (
	"context"
	"net/http"
	"strings"

	"SPOTCIN/backend/internal/services"
)

type userIDKey string

const UserIDKey userIDKey = "user_id"

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "token ausente", http.StatusUnauthorized)
			return
		}
		userID, err := services.ParseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			http.Error(w, "token inválido", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(UserIDKey).(string); ok {
		return v
	}
	return ""
}