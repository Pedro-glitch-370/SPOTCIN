package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"SPOTCIN/backend/internal/database"
	"SPOTCIN/backend/internal/models"
)

var ErrPlaybackNotFound = errors.New("posição não encontrada")

// playbackKey monta a chave no Redis
// Essa chave tem o formato playback:<user_id>:<track_id>
func playbackKey(userID, trackID string) string {
	return fmt.Sprintf("playback:%s:%s", userID, trackID)
}

// GetPlayback recupera a posição salva
// Se não existir, retorna ErrPlaybackNotFound
func GetPlayback(ctx context.Context, trackID, userID string) (*models.Playback, error) {
	val, err := database.RDB.Get(ctx, playbackKey(userID, trackID)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrPlaybackNotFound
		}
		return nil, err
	}

	var p models.Playback
	if err := json.Unmarshal([]byte(val), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// SavePlayback grava a posição com TTL de 1h
func SavePlayback(ctx context.Context, trackID, userID string, posicaoSeg float64) error {
	p := models.Playback{
		TrackID:      trackID,
		PosicaoSeg:   posicaoSeg,
		AtualizadoEm: time.Now().UTC(),
	}
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return database.RDB.Set(ctx, playbackKey(userID, trackID), data, time.Hour).Err()
}