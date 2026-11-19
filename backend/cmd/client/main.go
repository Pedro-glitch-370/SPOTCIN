package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func main() {
	var (
		base    = flag.String("base", "http://localhost:8080", "URL do backend")
		songID  = flag.String("song", "", "ID da música (se vazio, lista e pede escolha)")
		start   = flag.Float64("start", -1, "posição inicial em segundos (default: do Redis; -1 = automático)") //PONTODEATENCAO
		userID  = flag.String("user", "", "user_id (default: gera UUID v4)")
		minBuf  = flag.Int64("min", 64*1024, "buffer mínimo (bytes)")
		maxBuf  = flag.Int64("max", 512*1024, "buffer máximo (bytes)")
		jitter  = flag.Duration("jitter", 50*1000*1000, "jitter do consumo (ex.: 50ms)")
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

	// Passo 4 é abrir stream, montar o buffer, ligar o player, essas coisas
	// Mas, por ora, vamos só provar que as chamadas de API funcionam
	fmt.Printf("buffer: min=%dB max=%dB jitter=%s\n", *minBuf, *maxBuf, *jitter)

	fmt.Println("OK: cliente 6.1 inicializado (stream virá no 6.2)")
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