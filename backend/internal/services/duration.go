package services

import (
	"fmt"
	"io"
	"os"

	"github.com/go-audio/riff"
	"github.com/tcolgate/mp3"
)

// extractDuration retorna a duração em segundos pra MP3 e WAV
// Pra outros formatos, retorna 0, mas não bloqueia o scaner
func extractDuration(path, mime string) (int, error) {
	switch mime {
	case "audio/mpeg":
		return durationMP3(path)
	case "audio/wav":
		return durationWAV(path)
	default:
		return 0, nil
	}
}

// durationMP3 soma a duração de cada frame do MP3
func durationMP3(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	d := mp3.NewDecoder(f)
	var frame mp3.Frame
	skipped := 0
	total := 0.0

	for {
		if err := d.Decode(&frame, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
		total += frame.Duration().Seconds()
	}

	return int(total), nil
}

// durationWAV lê o header RIFF/WAV e calcula a duração
func durationWAV(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	d, err := riff.Duration(f)
	if err != nil {
		return 0, fmt.Errorf("riff duration: %w", err)
	}

	return int(d.Seconds()), nil
}