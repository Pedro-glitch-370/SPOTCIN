package services

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"SPOTCIN/backend/internal/models"

	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"gorm.io/gorm"
)

// AudioDir é o diretório local de áudio (que é temporário, vamos tirar depois).
var AudioDir string

func init() {
	dir := os.Getenv("AUDIO_DIR")
	if dir == "" {
		dir = "devseed/audio"
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("resolvendo AUDIO_DIR: %v", err)
	}
	AudioDir = abs
}

// Sincroniza AudioDir com o DB: insere novos, atualiza
// mudados e remove registros de arquivos que sumiram do disco
func ScanDatabase(db *gorm.DB) error {
	entries, err := os.ReadDir(AudioDir)
	if err != nil {
		return fmt.Errorf("lendo %s: %w", AudioDir, err)
	}

	// present guarda os nomes de áudio que existem no disco, pra depois
	// sabermos o que sobrou no DB sem arquivo correspondente
	present := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !isAudioExtension(entry.Name()) {
			continue
		}
		present[entry.Name()] = true

		if err := syncSong(db, entry); err != nil {
			log.Printf("%s: %v", entry.Name(), err)
		}
	}

	// remove do DB o que não está mais no disco
	if err := removeOrphans(db, present); err != nil {
		return fmt.Errorf("removendo órfãos: %w", err)
	}

	return nil
}

// syncSong insere ou atualiza uma única música no DB
func syncSong(db *gorm.DB, entry os.DirEntry) error {
	relPath := entry.Name()
	absPath := filepath.Join(AudioDir, relPath)

	mime, err := detectAudioType(absPath)
	if err != nil {
		return err
	}

	info, err := entry.Info()
	if err != nil {
		return err
	}

	duration, err := extractDuration(absPath, mime)
	if err != nil {
		log.Printf("duração de %s: %v (seguindo com 0)", relPath, err)
		duration = 0
	}

	var existing models.Song
	err = db.Where("path = ?", relPath).First(&existing).Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return db.Create(&models.Song{
			ID:       uuid.NewString(),
			Path:     relPath,
			Title:    humanizeName(relPath),
			Size:     info.Size(),
			ModTime:  info.ModTime(),
			MimeType: mime,
			Duration: duration,
		}).Error

	case err != nil:
		return err

	case existing.Size != info.Size() || !existing.ModTime.Equal(info.ModTime()):
		// arquivo mudou no disco, atualiza metadados.
		existing.Size = info.Size()
		existing.ModTime = info.ModTime()
		existing.MimeType = mime
		existing.Duration = duration
		return db.Save(&existing).Error

	default:
		return nil
	}
}

// Apaga do DB as músicas cujo path não está em present
func removeOrphans(db *gorm.DB, present map[string]bool) error {
	var songs []models.Song
	if err := db.Find(&songs).Error; err != nil {
		return err
	}

	for _, s := range songs {
		if present[s.Path] {
			continue
		}
		if err := db.Delete(&s).Error; err != nil {
			log.Printf("removendo %s: %v", s.Path, err)
			continue
		}
		log.Printf("removido do DB: %s", s.Path)
	}
	return nil
}

// isAudioExtension verifica a extensão do arquivo e retorna true
// apenas para formatos de áudio suportados
func isAudioExtension(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".wav", ".mp3", ".flac", ".ogg", ".m4a", ".aac", ".aiff":
		return true
	}
	return false
}

// Lê os primeiros bytes do arquivo e retorna o MIME
// type de áudio detectado, já normalizado
func detectAudioType(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// 261 bytes é o máximo que a lib filetype precisa pro header
	head := make([]byte, 261)
	n, err := f.Read(head)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	kind, err := filetype.Match(head[:n])
	if err != nil {
		return "", err
	}
	if kind == filetype.Unknown {
		return "", errors.New("tipo desconhecido")
	}
	if !strings.HasPrefix(kind.MIME.Value, "audio/") {
		return kind.MIME.Value, fmt.Errorf("não é áudio: %s", kind.MIME.Value)
	}
	return normalizeMime(kind.MIME.Value), nil
}

// normalizeMime converte MIMEs legados com prefixo "x-" pro padrão
// atual (ex: audio/x-wav -> audio/wav)
func normalizeMime(m string) string {
	switch m {
	case "audio/x-wav":
		return "audio/wav"
	case "audio/x-flac":
		return "audio/flac"
	case "audio/x-aiff":
		return "audio/aiff"
	}
	return m
}

// Transforma o nome do arquivo num título legível:
// remove a extensão, troca "_" e "-" por espaço e capitaliza cada palavra
func humanizeName(filename string) string {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	if base == "" {
		return base
	}

	base = strings.NewReplacer("_", " ", "-", " ").Replace(base)
	words := strings.Fields(base)
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}

	return strings.Join(words, " ")
}