package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func main() {
	var (
		base    = flag.String("base", "http://localhost:8080", "URL do backend")
		songID  = flag.String("song", "", "ID da música (se vazio, lista e pede escolha)")
		start   = flag.Float64("start", -1, "posição inicial em segundos")
		userID  = flag.String("user", "", "user_id (default: gera UUID v4)")
		minBuf  = flag.Int64("min", 64*1024, "buffer mínimo (bytes)")
		maxBuf  = flag.Int64("max", 512*1024, "buffer máximo (bytes)")
		jitter  = flag.Duration("jitter", 50*1000*1000, "jitter do consumo")
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
	player.posBytes = offset

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

	// downloader com loop de reconexão
	downloaderDone := make(chan struct{})
	go func() {
		defer close(downloaderDone)
		defer buf.Close() // sinaliza fim de arquivo pro player quando desistir

		totalDownloaded := int64(0)
		connections := 0
		for {
			// Se o player já acabou, não reconecta mais
			if player.PosBytes() >= meta.Size {
				log.Printf("downloader: player consumiu tudo (%d/%d bytes), encerrando",
					player.PosBytes(), meta.Size)
				return
			}

			// offset de retomada, onde o player tá
			resumeOffset := player.PosBytes()
			log.Printf("conexão #%d: resume em %d bytes (%.2fs)",
				connections+1, resumeOffset, BytesToPos(resumeOffset, meta))

			n, err := runDownloader(ctx, *base, id, resumeOffset, buf, *dropAfter, nil)
			connections++
			totalDownloaded += n

			// fim normal, sem drop
			if err == nil && *dropAfter == 0 {
				return 
			}

			// erro real
			if err != nil && err != io.EOF && ctx.Err() == nil {
				if strings.Contains(err.Error(), "416") {
					log.Printf("downloader: 416, fim do arquivo, encerrando")
					return
				}
				log.Printf("downloader: reconectando após erro: %v", err)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			// drop proposital, reconecta se ainda não consumiu tudo
			if *dropAfter > 0 {
				log.Printf("downloader: reconectando após drop (total baixado=%d KB)", totalDownloaded/1024)
				time.Sleep(200 * time.Millisecond)
				continue
			}
			return
		}
	}()

	// espera o player terminar o arquivo ou sair
	<-playerDone
	cancel()
	<-downloaderDone

	// PUT final
	finalPos := BytesToPos(player.PosBytes(), meta)
	api.PutPlayback(context.Background(), id, finalPos)
	fmt.Printf("fim: posição final %.2fs (%d bytes)\n", finalPos, player.PosBytes())
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
	buf *Buffer,
	dropAfter int64,
	onProgress func(total int64),
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
			buf.Write(chunk[:n])
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