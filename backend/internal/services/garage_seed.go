package services

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"

	"SPOTCIN/backend/internal/database"
)

// SeedGarage sobe pro bucket os arquivos de AudioDir que ainda não existem
func SeedGarage(ctx context.Context) error {
	entries, err := os.ReadDir(AudioDir)
	if err != nil {
		return err
	}

	uploaded, skipped := 0, 0
	for _, entry := range entries {
		if entry.IsDir() || !isAudioExtension(entry.Name()) {
			continue
		}
		key := entry.Name()
		abs := filepath.Join(AudioDir, key)

		if _, err := database.GarageClient.StatObject(ctx, database.GarageBucket, key, minio.StatObjectOptions{}); err == nil {
			skipped++
			continue
		}

		mime, err := detectAudioType(abs)
		if err != nil {
			slog.Warn("seed: tipo desconhecido", "file", key, "err", err)
			continue
		}

		_, err = database.GarageClient.FPutObject(ctx, database.GarageBucket, key, abs, minio.PutObjectOptions{
			ContentType: mime,
		})
		if err != nil {
			slog.Error("seed: falha no upload", "file", key, "err", err)
			continue
		}
		uploaded++
	}

	slog.Info("seed_garage", "uploaded", uploaded, "skipped", skipped, "bucket", database.GarageBucket)
	return nil
}

// helper pra não depender de strings
var _ = strings.HasPrefix