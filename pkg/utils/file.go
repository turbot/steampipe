package utils

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"time"
)

func FileModTime(filePath string) (time.Time, error) {
	file, err := os.Stat(filePath)

	if err != nil {
		return time.Time{}, err
	}

	return file.ModTime(), nil
}

// MoveFile moves a file from source to destiantion.
//
//	It first attempts the movement using OS primitives (os.Rename)
//	If os.Rename fails, it copies the file byte-by-byte to the destination and then removes the source
func MoveFile(source string, destination string) error {
	// try an os.Rename - it is always faster than copy
	err := os.Rename(source, destination)
	if err == nil {
		return nil
	}

	// os.Rename did not work.
	// do a byte-by-byte copy
	srcFile, err := os.OpenFile(source, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	dstFile, err := os.OpenFile(destination, os.O_WRONLY, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return os.Remove(source)
}

func FilenameNoExtension(fileName string) string {
	fileName = path.Base(fileName)
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}
