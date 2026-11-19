package main

import (
	"sync"
)

// Buffer é um buffer concorrente com nível mínimo e máximo
// O downloader escreve e o player lê
type Buffer struct {
	mu     sync.Mutex
	cond   *sync.Cond
	data   []byte
	min    int
	max    int
	closed bool
}

func NewBuffer(min, max int) *Buffer {
	b := &Buffer{min: min, max: max}
	b.cond = sync.NewCond(&b.mu)
	return b
}

// Write bloqueia enquanto o buffer estiver cheio
func (b *Buffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Bloqueia aqui, o downloader só lê mais do socket quando o player consumir
	for len(b.data) >= b.max && !b.closed {
		b.cond.Wait()
	}
	if b.closed {
		return 0, nil
	}
	b.data = append(b.data, p...)
	b.cond.Broadcast()
	return len(p), nil
}

// Read consome até n bytes e bloqueia enquanto o buffer estiver abaixo do mínimo
func (b *Buffer) Read(n int) []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	// O player só começa a ler quando o buffer atinge oo nível mínimo
	for len(b.data) < b.min && !b.closed {
		b.cond.Wait()
	}
	if b.closed && len(b.data) == 0 {
		return nil
	}
	if n > len(b.data) {
		n = len(b.data)
	}
	out := make([]byte, n)
	copy(out, b.data[:n])
	b.data = b.data[n:]
	b.cond.Broadcast()
	return out
}

// Len retorna o tamanho atual do buffer
func (b *Buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.data)
}

// Close sinaliza o fim do arquivo e acorda quem tava esperando
func (b *Buffer) Close() {
	b.mu.Lock()
	b.closed = true
	b.cond.Broadcast()
	b.mu.Unlock()
}