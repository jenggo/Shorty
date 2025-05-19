package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/gofiber/storage/minio"
	"github.com/gosimple/slug"
	"github.com/rs/zerolog/log"
)

var Storage *minio.Storage

func SlugifyFilename(filename string) string {
	// Common double extensions
	doubleExts := []string{".tar.gz", ".tar.bz2", ".tar.xz", ".tar.zst"}

	var ext string
	var nameWithoutExt string

	// Check for double extensions first
	hasDoubleExt := false
	for _, doubleExt := range doubleExts {
		if strings.HasSuffix(filename, doubleExt) {
			ext = doubleExt
			nameWithoutExt = strings.TrimSuffix(filename, ext)
			hasDoubleExt = true
			break
		}
	}

	// If no double extension found, handle as single extension
	if !hasDoubleExt {
		ext = filepath.Ext(filename)
		nameWithoutExt = strings.TrimSuffix(filename, ext)
	}

	return slug.MakeLang(nameWithoutExt, "en") + ext
}

func GenerateState() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Error().Caller().Err(err).Send()
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

// ToJSON converts a struct to JSON bytes
func ToJSON(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		log.Error().Caller().Err(err).Send()
		return nil
	}
	return data
}

// FromJSON unmarshals JSON data into a struct
func FromJSON(data []byte, v any) error {
	if err := json.Unmarshal(data, v); err != nil {
		log.Error().Caller().Err(err).Send()
		return err
	}
	return nil
}
