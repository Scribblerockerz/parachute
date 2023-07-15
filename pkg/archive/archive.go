package archive

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

func NewArchive(source []string, destination string) (string, error) {

	directory, fileName, err := buildPath(destination, "archive.zip")
	if err != nil {
		return "", err
	}

	targetFilePath := path.Join(directory, fileName)

	err = zipSource(source, targetFilePath)
	if err != nil {
		return "", err
	}

	return targetFilePath, nil
}

func buildPath(destination string, fileName string) (string, string, error) {
	var directory string
	var err error

	if destination == "" {
		directory, err = os.Getwd()

		if err != nil {
			return "", "", err
		}

		return directory, fileName, nil
	}

	info, err := os.Stat(destination)
	isExisting := !errors.Is(err, os.ErrNotExist)

	if err != nil && isExisting {
		return "", "", err
	}

	if info != nil && info.IsDir() {
		return destination, fileName, nil
	}

	directory, fileName = filepath.Split(destination)

	if filepath.Ext(fileName) == "" {
		fileName = fmt.Sprintf("%s.zip", fileName)
	}

	return directory, fileName, nil
}

func zipSource(source []string, target string) error {
	// 1. Create a ZIP file and zip.Writer
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()

	for i := range source {
		currentSourcePath := source[i]

		err := filepath.Walk(source[i], func(path string, info os.FileInfo, err error) error {
			return addPathToZip(writer, currentSourcePath, path)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func addPathToZip(writer *zip.Writer, basepath string, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Method = zip.Deflate

	header.Name, err = filepath.Rel(filepath.Dir(basepath), path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		header.Name += "/"
	}

	// 5. Create writer for the file header and save content of the file
	headerWriter, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(headerWriter, f)
	return err
}
