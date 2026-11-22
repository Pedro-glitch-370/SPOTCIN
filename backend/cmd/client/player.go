package main

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// Player simula a reprodução consumindo do buffer
type Player struct {
	buf      *Buffer
	bitrate  float64
	jitter   time.Duration // oscilação máxima de rede, é aleatório
	mu       sync.Mutex
	posBytes int64 // bytes consumidos (acaba sendo o offset atual)
	gen      int
	paused   bool
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

		if p.Pausado() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// Read espera o buffer atingir o nível mínimo de bytes antes de dar o primeiro chunk
		chunk, gen := p.buf.Read(chunkSize)
		if chunk == nil {
			// fim do arquivo
			onProgress(p.PosBytes())
			return
		}

		// simula o tempo de reprodução do chunk, dado por bytes/bitrate mais o jitter
		dur := time.Duration(float64(len(chunk)) / p.bitrate * float64(time.Second))
		if p.jitter > 0 {
			j := time.Duration(rand.Int63n(int64(p.jitter)*2)) - p.jitter
			dur += j
			if dur < 0 {
				dur = 0
			}
		}
		if !p.tocar(ctx, len(chunk), dur, gen) {
			return
		}

		// A cada 5 segundos de reprodução, faz o PUT
		if time.Since(lastProgress) >= 5*time.Second {
			onProgress(p.PosBytes())
			lastProgress = time.Now()
		}
	}
}

// PosBytes retorna o offset consumido atual
func (p *Player) PosBytes() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.posBytes
}

func (p *Player) tocar(ctx context.Context, n int, dur time.Duration, gen int) bool {
	const passo = 50 * time.Millisecond
	var tocado time.Duration
	var contados int64
	for {
		select {
		case <-ctx.Done():
			return false
		default:
		}

		p.mu.Lock()
		if gen != p.gen {
			p.mu.Unlock()
			return true
		}
		alvo := int64(n)
		if tocado < dur {
			alvo = int64(float64(n) * float64(tocado) / float64(dur))
		}
		p.posBytes += alvo - contados
		contados = alvo
		pausado := p.paused
		p.mu.Unlock()

		if contados == int64(n) {
			return true
		}
		if pausado {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		d := min(passo, dur-tocado)
		time.Sleep(d)
		if !p.Pausado() {
			tocado += d
		}
	}
}

func (p *Player) Posicionar(posBytes int64, gen int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.posBytes = posBytes
	p.gen = gen
}

func (p *Player) AlternarPausa() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.paused = !p.paused
	return p.paused
}

func (p *Player) Pausado() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.paused
}
