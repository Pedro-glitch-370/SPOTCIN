package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SongMeta é o retorno de GET /api/songs/{id}
type SongMeta struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Artist   string  `json:"artist"`
	Duration int     `json:"duration"`
	Size     int64   `json:"size"`
	MimeType string  `json:"mime_type"`
}

// Playback é o retorno de GET /api/playback/{id}
type Playback struct {
	TrackID      string    `json:"track_id"`
	PosicaoSeg   float64   `json:"posicao_seg"`
	AtualizadoEm time.Time `json:"atualizado_em"`
}

// APIClient encapsula as chamadas HTTP pro back
type APIClient struct {
	BaseURL string
	UserID  string
	HTTP    *http.Client
}

func NewAPIClient(baseURL, userID string) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		UserID:  userID,
		HTTP:    &http.Client{Timeout: 30 * time.Second}, //PONTODEATENCAO
	}
}

// ListSongs retorna o catálogo
func (c *APIClient) ListSongs(ctx context.Context) ([]SongMeta, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/songs", nil)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /api/songs: %s", resp.Status)
	}
	var songs []SongMeta
	if err := json.NewDecoder(resp.Body).Decode(&songs); err != nil {
		return nil, err
	}
	return songs, nil
}

// GetSong retorna metadados de uma faixa
func (c *APIClient) GetSong(ctx context.Context, id string) (*SongMeta, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/songs/"+id, nil)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("track_nao_encontrada")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /api/songs/%s: %s", id, resp.Status)
	}
	var s SongMeta
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

// GetPlayback retorna a posição salva
// ok=false se vier sem posição
func (c *APIClient) GetPlayback(ctx context.Context, id string) (*Playback, bool, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/playback/"+id, nil)
	req.Header.Set("X-User-ID", c.UserID)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("GET /api/playback/%s: %s", id, resp.Status)
	}
	var p Playback
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, false, err
	}
	return &p, true, nil
}

// PutPlayback grava a posição
func (c *APIClient) PutPlayback(ctx context.Context, id string, posicaoSeg float64) error {
	body, _ := json.Marshal(map[string]float64{"posicao_seg": posicaoSeg})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, c.BaseURL+"/api/playback/"+id, bytes.NewReader(body))
	req.Header.Set("X-User-ID", c.UserID)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("PUT /api/playback/%s: %s", id, resp.Status)
	}
	return nil
}

// OpenStream faz GET /api/stream/{id} com Range a partir de offsetBytes
func (c *APIClient) OpenStream(ctx context.Context, id string, offsetBytes int64) (io.ReadCloser, int, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/stream/"+id, nil)
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offsetBytes))
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, err
	}
	// Devolve o body junto com status (206, 200 ou 416) pra main poder diferenciar
	return resp.Body, resp.StatusCode, nil
}

// O back não faz nenhuma das duas conversões abaixo, elas só são feitas no cliente

// PosToBytes converte posicao_seg do Redis pra offset em bytes usando a taxa de bits
func PosToBytes(posicaoSeg float64, meta *SongMeta) int64 {
	if meta.Duration <= 0 || meta.Size <= 0 {
		return 0
	}
	bitrate := float64(meta.Size) / float64(meta.Duration) // bytes (size) por segundo (duration)
	off := int64(posicaoSeg * bitrate)
	if off < 0 {
		return 0
	}
	if off > meta.Size {
		return meta.Size
	}
	return off
}

// BytesToPos converte offset em bytes pra posicao_seg
func BytesToPos(offsetBytes int64, meta *SongMeta) float64 {
	if meta.Duration <= 0 || meta.Size <= 0 {
		return 0
	}
	bitrate := float64(meta.Size) / float64(meta.Duration)
	return float64(offsetBytes) / bitrate
}