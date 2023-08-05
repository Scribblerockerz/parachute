package archive

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const DEFAULT_FILE_PERMISSIONS = 0755

type Archive struct {
	TempLocation string
	IsEncrupted  bool

	fileName string
}

func (a *Archive) TempDestination() string {
	if a.IsEncrupted {
		return a.encZipDestination()
	}
	return a.zipDestination()
}

func (a *Archive) Destination() string {
	return path.Join(a.TempLocation, a.fileName)
}

func (a *Archive) zipDestination() string {
	return path.Join(a.TempLocation, fmt.Sprintf("%s%s", a.fileName, ".zip"))
}

func (a *Archive) encZipDestination() string {
	return path.Join(a.TempLocation, fmt.Sprintf("%s%s", a.fileName, ".zip.enc"))
}

func (a *Archive) Zip(sources []string) error {
	return ZipSource(sources, a.zipDestination())
}

func (a *Archive) Unzip() error {
	return UnzipSource(a.zipDestination(), a.Destination())
}

func (a *Archive) Encrypt(passphrase string) error {
	if passphrase == "" {
		log.Warn().Msg("provided passphrase is empty")
	}

	return EncryptFile(a.zipDestination(), a.encZipDestination(), passphrase)
}

func (a *Archive) Decrypt(passphrase string) error {
	if passphrase == "" {
		log.Warn().Msg("provided passphrase is empty")
	}

	return DecryptFile(a.encZipDestination(), a.zipDestination(), passphrase)
}

func (a *Archive) Cleanup() error {
	if a.IsEncrupted {
		err := os.Remove(a.encZipDestination())
		if err != nil {
			return err
		}

		log.Debug().Str("encryptedArchive", a.encZipDestination()).Msg("removed temporary encrypted archive")
	}

	err := os.Remove(a.zipDestination())
	if err != nil {
		return err
	}

	log.Debug().Str("archive", a.zipDestination()).Msg("removed temporary archive")

	return nil
}

func (a *Archive) CopyIntoDir(source string, destination string, useTimedName bool) (string, error) {
	destination, err := ensureValidDestination(destination)
	if err != nil {
		return "", err
	}

	fileName := path.Base(source)
	if useTimedName {
		fileName = fmt.Sprintf("%s_%s", time.Now().Format("20060102150405"), fileName)
	}

	destinationFilePath := path.Join(destination, fileName)

	if IsDir(source) {
		err = CopyDir(source, destinationFilePath)
	} else {
		err = CopyFile(source, destinationFilePath)
	}

	if err != nil {
		return "", err
	}

	log.Debug().Str("destination", destinationFilePath).Msg("created archive in target destination")

	return destinationFilePath, nil
}

func ensureValidDestination(destination string) (string, error) {
	var err error

	if destination == "" {
		destination, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	fileInfo, err := os.Stat(destination)
	notExists := errors.Is(err, os.ErrNotExist)
	if err != nil && !notExists {
		return "", err
	}

	if notExists {
		os.Mkdir(destination, DEFAULT_FILE_PERMISSIONS)
	}

	if !notExists && !fileInfo.IsDir() {
		return "", fmt.Errorf("provided output '%s' is not a directory", destination)
	}

	return destination, nil
}

func CreateArchiveFromSources(sources []string, useEncryption bool, passphrase string) (*Archive, error) {
	tmp, err := tempLocation()

	if err != nil {
		return nil, err
	}

	var fileName string

	if len(sources) == 1 {
		fileName = path.Base(sources[0])
	} else {
		fileName = "package"
	}

	if err != nil {
		return nil, err
	}

	a := &Archive{
		TempLocation: tmp,
		IsEncrupted:  useEncryption,
		fileName:     fileName,
	}

	err = a.Zip(sources)
	if err != err {
		return nil, err
	}

	log.Debug().Str("archive", a.zipDestination()).Msg("created temporary archive")

	if a.IsEncrupted {
		err = a.Encrypt(passphrase)

		if err != err {
			return nil, err
		}

		log.Debug().Str("encryptedArchive", a.encZipDestination()).Msg("encrypted temporary archive")
	}

	return a, nil
}

func CreateTempArchiveFromRemoteFile(remoteObjectPath string) (*Archive, error) {
	tmp, err := tempLocation()

	if err != nil {
		return nil, err
	}

	fileName := path.Base(remoteObjectPath)
	isEncrypted := strings.HasSuffix(fileName, ".enc")

	fileName, _ = strings.CutSuffix(fileName, ".enc")
	fileName, _ = strings.CutSuffix(fileName, ".zip")

	if err != nil {
		return nil, err
	}

	return &Archive{
		TempLocation: tmp,
		IsEncrupted:  isEncrypted,
		fileName:     fileName,
	}, nil
}

func tempLocation() (string, error) {
	return os.MkdirTemp("/tmp", "parachute")
}
