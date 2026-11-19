package main

import (
	"context"
	"math/rand"
	"time"
)

// Player simula a reprodução consumindo do buffer
type Player struct {
	buf      *Buffer
	bitrate  float64
	jitter   time.Duration // oscilação máxima de rede, é aleatório
	posBytes int64         // bytes consumidos (acaba sendo o offset atual)
}

func NewPlayer(buf *Buffer, bitrate float64, jitter time.Duration) *Player {
	return &Player{buf: buf, bitrate: bitrate, jitter: jitter}
}

// Run consome do buffer até dar fim de arquivo ou context cancelado
func (p *Player) Run(ctx context.Context, onProgress func(posBytes int64)) {
	// chunk com cerca de 1 segundo de áudio
	chunkSize := int(p.bitrate)
	if chunkSize < 4096 {
		chunkSize = 4096
	}

	lastProgress := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		// Read espera o buffer atingir o nível mínimo de bytes antes de dar o primeiro chunk
		chunk := p.buf.Read(chunkSize)
		if chunk == nil {
			// fim do arquivo
			onProgress(p.posBytes)
			return
		}

		// simula o tempo de reprodução do chunk, dado por bytes/bitrate mais o jitter
		dur := time.Duration(float64(len(chunk))/p.bitrate*float64(time.Second))
		if p.jitter > 0 {
			j := time.Duration(rand.Int63n(int64(p.jitter)*2)) - p.jitter
			dur += j
			if dur < 0 {
				dur = 0
			}
		}
		time.Sleep(dur)

		p.posBytes += int64(len(chunk))

		// A cada 5 segundos de reprodução, faz o PUT
		if time.Since(lastProgress) >= 5*time.Second {
			onProgress(p.posBytes)
			lastProgress = time.Now()
		}
	}
}

// PosBytes retorna o offset consumido atual
func (p *Player) PosBytes() int64 {
	return p.posBytes
}