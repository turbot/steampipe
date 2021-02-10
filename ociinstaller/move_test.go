package ociinstaller

import (
	"crypto/rand"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type DirectoryStructure struct {
	Name        string
	Files       map[string]string
	Directories map[string]DirectoryStructure
	// using maps here so that we may use DeepEqual for checking
}

func TestMoveFile(t *testing.T) {
	dir, err := setupForTest(t)
	if err != nil {
		t.Fatal(err)
		return
	}
	dstFile := filepath.Join(os.TempDir(), "dst")

	defer func() {
		os.RemoveAll(dir)
		os.Remove(dstFile)
	}()

	files, err := ioutil.ReadDir(dir)
	transferingFile := files[0]
	for _, f := range files {
		transferingFile = f
		if !transferingFile.IsDir() {
			break
		}
	}

	if err = moveFile(filepath.Join(dir, transferingFile.Name()), dstFile); err != nil {
		t.Fatal(err)
	}

	dstStat, _ := os.Stat(dstFile)

	if dstStat.Size() != transferingFile.Size() {
		t.Error("Incorrect size")
	}

	// now try to get the stat of the old file
	_, err = os.Stat(filepath.Join(dir, transferingFile.Name()))
	if err == nil {
		t.Error("Did not remove")
	}
}

func TestMoveFolder(t *testing.T) {
	srcFolder, err := setupForTest(t)
	dstFolder := filepath.Join(os.TempDir(), "dst")
	if err != nil {
		t.Fatal(err)
		return
	}

	defer func() {
		os.RemoveAll(srcFolder)
		os.RemoveAll(dstFolder)
	}()

	srcStruct, _ := getDirectoryStructureFrom(srcFolder)
	if err = moveFolder(srcFolder, dstFolder); err != nil {
		t.Error("failed to move", err)
		return
	}
	dstStruct, _ := getDirectoryStructureFrom(dstFolder)

	if !reflect.DeepEqual(srcStruct.Files, dstStruct.Files) {
		t.Error("Not equal")
	}

	_, err = os.Stat(srcFolder)
	if err == nil {
		t.Error("did not delete")
	}
}

func TestCopyFile(t *testing.T) {
	dir, err := setupForTest(t)
	if err != nil {
		t.Fatal(err)
		return
	}
	dstFile := filepath.Join(os.TempDir(), "dst")

	defer func() {
		os.RemoveAll(dir)
		os.Remove(dstFile)
	}()

	files, err := ioutil.ReadDir(dir)
	transferingFile := files[0]
	for _, f := range files {
		transferingFile = f
		if !transferingFile.IsDir() {
			break
		}
	}

	if err = copyFile(filepath.Join(dir, transferingFile.Name()), dstFile); err != nil {
		t.Fatal(err)
	}

	dstStat, _ := os.Stat(dstFile)

	if dstStat.Size() != transferingFile.Size() {
		t.Error("Incorrect size")
	}
}

func TestCopyFolder(t *testing.T) {
	srcFolder, err := setupForTest(t)
	dstFolder := filepath.Join(os.TempDir(), "dst")
	if err != nil {
		t.Fatal(err)
		return
	}

	defer func() {
		os.RemoveAll(srcFolder)
		os.RemoveAll(dstFolder)
	}()

	if err = copyFolder(srcFolder, dstFolder); err != nil {
		t.Error("failed to copy", err)
		return
	}
	srcStruct, _ := getDirectoryStructureFrom(srcFolder)
	dstStruct, _ := getDirectoryStructureFrom(dstFolder)

	if !reflect.DeepEqual(srcStruct.Files, dstStruct.Files) {
		t.Error("Not equal")
	}
}

// creates a temporary directory with a 100 1MB files within it!
func setupForTest(t *testing.T) (string, error) {
	srcDirectory, _ := ioutil.TempDir(os.TempDir(), "src")
	fileName := "file"

	if err := os.MkdirAll(srcDirectory, 0777); err != nil {
		return "", err
	}

	// generate 1MB content
	content := make([]byte, 1*1024*1024)
	rand.Read(content)

	for i := 0; i < 100; i++ {
		// create a file in source directory
		f, _ := ioutil.TempFile(srcDirectory, fileName)
		f.Write(content)
		f.Close()
	}

	// create a file in source directory
	return srcDirectory, nil
}

func getDirectoryStructureFrom(root string) (DirectoryStructure, error) {
	structure := DirectoryStructure{
		Name:        root,
		Files:       map[string]string{},
		Directories: map[string]DirectoryStructure{},
	}

	contents, _ := ioutil.ReadDir(root)

	for _, info := range contents {
		if info.IsDir() {
			s, e := getDirectoryStructureFrom(filepath.Join(root, info.Name()))
			if e != nil {
				return structure, e
			}
			structure.Directories[info.Name()] = s
		} else {
			structure.Files[info.Name()] = info.Name()
		}
	}

	return structure, nil
}
