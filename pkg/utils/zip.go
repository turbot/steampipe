package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Unzip - from https://stackoverflow.com/questions/20357223/easy-way-to-unzip-file-with-golang
func Unzip(src, dst string) ([]string, error) {
	var files []string
	r, err := zip.OpenReader(src)
	if err != nil {
		return files, err
	}
	defer r.Close()

	err = os.MkdirAll(dst, 0755)
	if err != nil {
		return files, err
	}

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dst, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dst)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(path, f.Mode())
			if err != nil {
				return err
			}
		} else {
			err = os.MkdirAll(filepath.Dir(path), 0777)
			if err != nil {
				return err
			}
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
			if err != nil {
				return err
			}

			if _, err = io.Copy(f, rc); err != nil {
				return err
			}
			f.Close()
		}

		return nil
	}

	for _, f := range r.File {
		if err := extractAndWriteFile(f); err != nil {
			return files, err
		}
		files = append(files, f.FileHeader.Name)
	}

	return files, nil
}
