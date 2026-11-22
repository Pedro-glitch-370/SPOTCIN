package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const ajudaComandos = "comandos: p (pausar/retomar) | a <seg> (avançar) | v <seg> (voltar) | i <seg> (ir para) | q (sair)"

func main() {
	var (
		base      = flag.String("base", "http://localhost:8080", "URL do backend")
		songID    = flag.String("song", "", "ID da música (se vazio, lista e pede escolha)")
		start     = flag.Float64("start", -1, "posição inicial em segundos")
		userID    = flag.String("user", "", "user_id (default: gera UUID v4)")
		minBuf    = flag.Int64("min", 64*1024, "buffer mínimo (bytes)")
		maxBuf    = flag.Int64("max", 512*1024, "buffer máximo (bytes)")
		jitter    = flag.Duration("jitter", 50*1000*1000, "jitter do consumo")
		dropAfter = flag.Int64("drop-after", 0, "fecha o socket após N bytes baixados")
	)
	flag.Parse()

	if *userID == "" {
		*userID = uuid.NewString()
	}

	ctx := context.Background()
	api := NewAPIClient(*base, *userID)

	fmt.Printf("cliente: user_id=%s base=%s\n", *userID, *base)

	// Passo 1 é escolher uma música
	id := *songID
	if id == "" {
		id = escolherMusica(ctx, api)
		if id == "" {
			log.Fatal("nenhuma música escolhida")
		}
	}

	// Passo 2 é ter os metadados
	meta, err := api.GetSong(ctx, id)
	if err != nil {
		log.Fatalf("metadados: %v", err)
	}
	if meta.Duration <= 0 || meta.Size <= 0 {
		log.Fatalf("metadados sem duração ou tamanho válidos (duration=%d size=%d)", meta.Duration, meta.Size)
	}
	bitrate := float64(meta.Size) / float64(meta.Duration)
	fmt.Printf("faixa: %s (%s) duration=%ds size=%dB bitrate=%.0f B/s\n",
		meta.Title, meta.MimeType, meta.Duration, meta.Size, bitrate)

	// Passo 3 é ter a posição inicial
	var posicaoSeg float64
	if *start >= 0 {
		posicaoSeg = *start
		fmt.Printf("posição inicial: %.2fs (flag -start)\n", posicaoSeg)
	} else {
		p, ok, err := api.GetPlayback(ctx, id)
		if err != nil {
			log.Fatalf("playback GET: %v", err)
		}
		if ok {
			posicaoSeg = p.PosicaoSeg
			fmt.Printf("posição inicial: %.2fs (Redis, atualizado em %s)\n", posicaoSeg, p.AtualizadoEm.Format("15:04:05"))
		} else {
			posicaoSeg = 0
			fmt.Printf("posição inicial: 0s (sem posição salva)\n")
		}
	}

	offset := PosToBytes(posicaoSeg, meta)
	fmt.Printf("offset inicial: %d bytes\n", offset)

	// Passo 4 é buffer + player
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	buf := NewBuffer(int(*minBuf), int(*maxBuf))
	player := NewPlayer(buf, bitrate, *jitter)
	player.Posicionar(offset, 0)

	ctrl := &controle{
		ctx:       ctx,
		base:      *base,
		id:        id,
		meta:      meta,
		buf:       buf,
		player:    player,
		dropAfter: *dropAfter,
	}
	ctrl.iniciar(offset, 0)

	// o player é em goroutine
	playerDone := make(chan struct{})
	go func() {
		player.Run(ctx, func(posBytes int64) {
			posSeg := BytesToPos(posBytes, meta)
			if err := api.PutPlayback(ctx, id, posSeg); err != nil {
				log.Printf("PUT playback: %v", err)
			} else {
				log.Printf("PUT playback: pos=%.2fs", posSeg)
			}
		})
		close(playerDone)
	}()

	fmt.Println(ajudaComandos)
	go lerComandos(ctrl, api, cancel)

	// espera o player terminar o arquivo ou sair
	<-playerDone
	ctrl.parar()
	cancel()

	// PUT final
	finalPos := BytesToPos(player.PosBytes(), meta)
	api.PutPlayback(context.Background(), id, finalPos)
	fmt.Printf("fim: posição final %.2fs (%d bytes)\n", finalPos, player.PosBytes())
}

type controle struct {
	mu        sync.Mutex
	ctx       context.Context
	base      string
	id        string
	meta      *SongMeta
	buf       *Buffer
	player    *Player
	dropAfter int64
	cancelar  context.CancelFunc
	feito     <-chan struct{}
	encerrado bool
}

func (c *controle) iniciar(offset int64, gen int) {
	c.cancelar, c.feito = iniciarDownload(c.ctx, c.base, c.id, offset, gen, c.meta, c.buf, c.dropAfter)
}

func (c *controle) irPara(posSeg float64) (int64, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.encerrado {
		return 0, false
	}
	off := PosToBytes(posSeg, c.meta)
	c.cancelar()
	gen := c.buf.Reiniciar()
	<-c.feito
	c.player.Posicionar(off, gen)
	c.iniciar(off, gen)
	return off, true
}

func (c *controle) parar() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.encerrado = true
	c.cancelar()
	c.buf.Close()
	<-c.feito
}

func lerComandos(c *controle, api *APIClient, sair func()) {
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		campos := strings.Fields(strings.ToLower(sc.Text()))
		if len(campos) == 0 {
			continue
		}
		atual := BytesToPos(c.player.PosBytes(), c.meta)

		switch campos[0] {
		case "p":
			if c.player.AlternarPausa() {
				fmt.Printf("pausado em %.2fs\n", atual)
			} else {
				fmt.Printf("retomado em %.2fs\n", atual)
			}

		case "a", "v", "i":
			if len(campos) < 2 {
				fmt.Println("uso: a <seg> | v <seg> | i <seg>")
				continue
			}
			seg, err := strconv.ParseFloat(campos[1], 64)
			if err != nil || seg < 0 {
				fmt.Println("valor inválido")
				continue
			}
			alvo := seg
			switch campos[0] {
			case "a":
				alvo = atual + seg
			case "v":
				alvo = atual - seg
			}
			alvo = math.Max(0, math.Min(alvo, float64(c.meta.Duration)))

			off, ok := c.irPara(alvo)
			if !ok {
				return
			}
			fmt.Printf("posição: %.2fs (Range: bytes=%d-)\n", alvo, off)
			if err := api.PutPlayback(context.Background(), c.id, alvo); err != nil {
				log.Printf("PUT playback: %v", err)
			}

		case "q":
			sair()
			return

		default:
			fmt.Println(ajudaComandos)
		}
	}
}

func iniciarDownload(
	parent context.Context,
	base, id string,
	offset int64,
	gen int,
	meta *SongMeta,
	buf *Buffer,
	dropAfter int64,
) (context.CancelFunc, <-chan struct{}) {
	ctx, cancel := context.WithCancel(parent)
	done := make(chan struct{})

	go func() {
		defer close(done)

		next := offset
		connections := 0
		for {
			if next >= meta.Size {
				log.Printf("downloader: arquivo completo (%d/%d bytes), encerrando", next, meta.Size)
				buf.Finalizar(gen)
				return
			}

			connections++
			log.Printf("conexão #%d: resume em %d bytes (%.2fs)",
				connections, next, BytesToPos(next, meta))

			n, err := runDownloader(ctx, base, id, next, gen, buf, dropAfter)
			next += n

			if ctx.Err() != nil || errors.Is(err, errBufferDescartado) {
				return
			}

			espera := 200 * time.Millisecond
			if err != nil {
				if strings.Contains(err.Error(), "416") {
					log.Printf("downloader: 416, fim do arquivo, encerrando")
					buf.Finalizar(gen)
					return
				}
				log.Printf("downloader: reconectando após erro: %v", err)
				espera = 500 * time.Millisecond
			} else if next < meta.Size {
				log.Printf("downloader: reconectando após drop (próximo offset=%d)", next)
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(espera):
			}
		}
	}()

	return cancel, done
}

// escolherMusica lista os áudios e lê o input do usuário
// Se o usuário não escolher nada, retorna ""
func escolherMusica(ctx context.Context, api *APIClient) string {
	songs, err := api.ListSongs(ctx)
	if err != nil {
		log.Fatalf("catálogo: %v", err)
	}
	if len(songs) == 0 {
		fmt.Println("catálogo vazio")
		return ""
	}

	fmt.Println("\nMúsicas disponíveis:")
	for i, s := range songs {
		fmt.Printf("  [%d] %s — %s (%ds, %s)\n", i+1, s.Title, s.Artist, s.Duration, s.MimeType)
	}

	sc := bufio.NewScanner(os.Stdin)
	defer func() {
		if err := sc.Err(); err != nil {
			log.Printf("stdin: %v", err)
		}
	}()
	for {
		fmt.Printf("\nEscolha (1-%d) ou cole um ID: ", len(songs))
		if !sc.Scan() {
			return ""
		}
		in := strings.TrimSpace(sc.Text())
		if in == "" {
			return ""
		}
		// Se o usuário tiver colado um ID
		if strings.Contains(in, "-") && len(in) > 10 {
			return in
		}
		// Se o usuário tiver escolhido um número
		n, err := strconv.Atoi(in)
		if err == nil && n >= 1 && n <= len(songs) {
			return songs[n-1].ID
		}
		fmt.Println("entrada inválida")
	}
}

// runDownloader abre o stream a partir de offsetBytes, lê e escreve no buffer
// Retorna os bytes que foram baixados nessa conexão
func runDownloader(
	ctx context.Context,
	base, id string,
	offsetBytes int64,
	gen int,
	buf *Buffer,
	dropAfter int64,
) (int64, error) {
	client := &http.Client{}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
		base+"/api/stream/"+id, nil)
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offsetBytes))

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusRequestedRangeNotSatisfiable {
		return 0, fmt.Errorf("offset %d fora do arquivo (416)", offsetBytes)
	}
	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("stream: %s", resp.Status)
	}

	log.Printf("stream aberto: status=%d range=bytes=%d-", resp.StatusCode, offsetBytes)

	chunk := make([]byte, 32*1024)
	var total int64
	for {
		select {
		case <-ctx.Done():
			return total, ctx.Err()
		default:
		}

		n, err := resp.Body.Read(chunk)
		if n > 0 {
			if _, werr := buf.Write(chunk[:n], gen); werr != nil {
				return total, werr
			}
			total += int64(n)
			if total%(1<<20) == 0 {
				log.Printf("downloader: %d KB baixados (conexão), buffer=%d KB", total/1024, buf.Len()/1024)
			}
			// fecha o corpo depois de N bytes baixados e retorna (só pra simular a queda)
			if dropAfter > 0 && total >= dropAfter {
				log.Printf("downloader: drop-after atingido (%d KB) — fechando socket", total/1024)
				return total, nil
			}
		}
		if err != nil {
			if err == io.EOF {
				log.Printf("downloader: fim natural (%d KB)", total/1024)
				return total, nil
			}
			log.Printf("downloader: erro (%d KB, %v)", total/1024, err)
			return total, err
		}
	}
}
