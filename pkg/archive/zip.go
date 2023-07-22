package archive

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ZipSource(source []string, target string) error {
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

// Unzip source: https://stackoverflow.com/a/24792688
func UnzipSource(source, target string) error {
	r, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	_, err = os.Stat(target)
	isNotExist := errors.Is(err, os.ErrNotExist)
	if err != nil && !isNotExist {
		return err
	}

	if !isNotExist {
		os.RemoveAll(target)
	}

	os.MkdirAll(target, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		// Keeps crashing with "./"
		if f.Name == "./" {
			return nil
		}

		path := filepath.Join(target, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(target)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
