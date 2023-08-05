package archive

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	cp "github.com/otiai10/copy"
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

// Source: https://stackoverflow.com/a/50741908
func CopyFile(source string, destination string) error {
	inputFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("unable to open source file: %s", err)
	}

	outputFile, err := os.Create(destination)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("unable to open destination file: %s", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}

	return nil
}

func CopyDir(source string, destination string) error {
	return cp.Copy(source, destination)
}

func IsDir(somePath string) bool {
	info, err := os.Stat(somePath)

	if err != nil {
		return false
	}

	return info.IsDir()
}
