# ---------- Stage 1: build ----------
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Cache de dependências
COPY go.mod go.sum ./
RUN go mod download

# Copia o código e builda binário estático
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/server ./cmd/server

# ---------- Stage 2: runtime ----------
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copia só o binário (imagem final ~5MB)
COPY --from=builder /app/server /app/server

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/server"]