package ociinstaller

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func ungzip(sourceFile string, destDir string) (string, error) {
	r, err := os.Open(sourceFile)
	if err != nil {
		return "", err
	}

	uncompressedStream, err := gzip.NewReader(r)
	if err != nil {
		return "", err
	}

	destFile := filepath.Join(destDir, uncompressedStream.Name)
	outFile, err := os.OpenFile(destFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(outFile, uncompressedStream); err != nil {
		return "", err
	}

	outFile.Close()
	if err := uncompressedStream.Close(); err != nil {
		return "", err
	}

	return destFile, nil
}

func fileExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func copyFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer outputFile.Close()

	if _, err = io.Copy(outputFile, inputFile); err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}

	// copy over the permissions and modes
	inputStat, _ := os.Stat(sourcePath)
	outputFile.Chmod(inputStat.Mode())

	return nil
}

func copyFileUnlessExists(sourcePath string, destPath string) error {
	if fileExists(destPath) {
		return nil
	}
	return copyFile(sourcePath, destPath)
}

// moves a file within an fs partition. panics if movement is attempted between partitions
// this is done separately to achieve performance benefits of os.Rename over reading and writing content
func moveFileWithinPartition(sourcePath, destPath string) error {
	if err := os.Rename(sourcePath, destPath); err != nil {
		return fmt.Errorf("error moving file: %s", err)
	}
	return nil
}

// moves a folder within an fs partition. panics if movement is attempted between partitions
// this is done separately to achieve performance benefits of os.Rename over reading and writing content
func moveFolderWithinPartition(sourcePath, destPath string) error {
	sourceinfo, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(destPath, sourceinfo.Mode()); err != nil {
		return err
	}

	directory, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("couldn't open source dir: %s", err)
	}
	directory.Close()

	defer os.RemoveAll(sourcePath)

	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		relPath, _ := filepath.Rel(sourcePath, path)
		if relPath == "" {
			return nil
		}
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(destPath, relPath), info.Mode())
		}
		return moveFileWithinPartition(filepath.Join(sourcePath, relPath), filepath.Join(destPath, relPath))
	})
}
