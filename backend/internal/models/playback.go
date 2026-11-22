package models

import "time"

// Playback representa a posição de reprodução de uma faixa por um usuário
// No Redis, ele é serializado como JSON pra depois ser devolvido na resposta HTTP
type Playback struct {
	TrackID      string    `json:"track_id"`
	PosicaoSeg   float64   `json:"posicao_seg"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}