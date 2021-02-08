package ociinstaller

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkMoveFileByRename(b *testing.B) {
	b.StopTimer()
	srcFolder, dstFolder, err := setup(b)
	defer func() {
		os.RemoveAll(srcFolder)
		os.RemoveAll(dstFolder)
	}()
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()
	if err := moveDir(srcFolder, dstFolder, true); err != nil {
		b.Fatal(err)
	}
}
func BenchmarkMoveFileByCopy(b *testing.B) {
	b.StopTimer()
	srcFolder, dstFolder, err := setup(b)
	defer func() {
		os.RemoveAll(srcFolder)
		os.RemoveAll(dstFolder)
	}()
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()
	if err := moveDir(srcFolder, dstFolder, false); err != nil {
		b.Fatal(err)
	}
}

func move(src, dest string, useRename bool) error {
	if useRename {
		return os.Rename(src, dest)
	}

	inputFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer outputFile.Close()

	if _, err = io.Copy(outputFile, inputFile); err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}
	return nil
}

func moveDir(source string, dest string, useRename bool) (err error) {
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(dest, sourceinfo.Mode()); err != nil {
		return err
	}

	directory, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("couldn't open source dir: %s", err)
	}
	defer directory.Close()

	objects, err := directory.Readdir(-1)
	if err != nil {
		return err
	}

	for _, obj := range objects {
		sourceFile := filepath.Join(source, obj.Name())
		destFile := filepath.Join(dest, obj.Name())

		if err := move(sourceFile, destFile, useRename); err != nil {
			return fmt.Errorf("error moving file: %s", err)
		}
	}
	return nil
}

func setup(b *testing.B) (string, string, error) {
	srcDirectory, _ := ioutil.TempDir(os.TempDir(), "")
	dstDirectory, _ := ioutil.TempDir(os.TempDir(), "")
	fileName := "file"

	if err := os.MkdirAll(srcDirectory, 0777); err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(dstDirectory, 0777); err != nil {
		return "", "", err
	}

	// generate 1MB content
	content := make([]byte, 1*1024*1024)
	rand.Read(content)

	for i := 0; i < b.N; i++ {
		// create a file in source directory
		f, _ := ioutil.TempFile(srcDirectory, fileName)
		f.Write(content)
		f.Close()
	}

	// create a file in source directory
	return srcDirectory, dstDirectory, nil
}
