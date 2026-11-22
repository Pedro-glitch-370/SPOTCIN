package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// Logging registra cada requisição como uma linha JSON
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Milliseconds()

		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rw.status),
			slog.Int64("duration_ms", duration),
			slog.Int("bytes", rw.bytes),
			slog.String("request_id", RequestIDFromContext(r.Context())),
		}

		if rg := r.Header.Get("Range"); rg != "" {
			attrs = append(attrs, slog.String("range", rg))
		}
		if cl := rw.Header().Get("Content-Length"); cl != "" {
			attrs = append(attrs, slog.String("content_length", cl))
		}
		if cr := rw.Header().Get("Content-Range"); cr != "" {
			attrs = append(attrs, slog.String("content_range", cr))
		}

		level := slog.LevelInfo
		switch {
		case rw.status >= 500:
			level = slog.LevelError
		case rw.status >= 400:
			level = slog.LevelWarn
		}

		slog.LogAttrs(r.Context(), level, "http_request", attrs...)
	})
}