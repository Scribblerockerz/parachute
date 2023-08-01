package archive

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const ENCRYPTED_FILE_SUFFIX = ".enc"

func GetFallbackName(hint string, fallback string, suffix string) string {
	if hint == "" {
		return fallback
	}

	basename := filepath.Base(hint)
	basename, _ = strings.CutSuffix(basename, suffix)

	return basename
}

func DestinationPath(destination string, fallbackFilename string, incrementIfExists bool) (string, error) {
	if destination == "" {
		cwd, err := os.Getwd()

		if err != nil {
			return "", err
		}

		destination = path.Join(cwd, fallbackFilename)
	}

	info, err := os.Stat(destination)

	if errors.Is(err, os.ErrNotExist) {
		return destination, nil
	}

	if info.IsDir() {
		destination = path.Join(destination, fallbackFilename)
	}

	_, err = os.Stat(destination)

	if errors.Is(err, os.ErrNotExist) {
		return destination, nil
	}

	if !incrementIfExists {
		return "", fmt.Errorf("destination '%s' does already exist", destination)
	}

	return fmt.Sprintf("%s%s", destination, time.Now().Format("20060102150405")), nil

}

func IsFileEncrypted(path string) bool {
	return strings.HasSuffix(path, ENCRYPTED_FILE_SUFFIX)
}
