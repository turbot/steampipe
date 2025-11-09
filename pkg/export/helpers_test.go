package export

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// errorReader simulates a reader that fails after some data is written
type errorReader struct {
	data      []byte
	position  int
	failAfter int
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	if e.position >= e.failAfter {
		return 0, errors.New("simulated write error")
	}

	remaining := e.failAfter - e.position
	toRead := len(p)
	if toRead > remaining {
		toRead = remaining
	}
	if toRead > len(e.data)-e.position {
		toRead = len(e.data) - e.position
	}

	if toRead == 0 {
		return 0, io.EOF
	}

	copy(p, e.data[e.position:e.position+toRead])
	e.position += toRead
	return toRead, nil
}

// TestWrite_PartialFileCleanup tests that Write() does not leave partial files
// when a write operation fails midway through.
// This test documents the expected behavior for bug #4718.
func TestWrite_PartialFileCleanup(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	targetFile := filepath.Join(tmpDir, "output.txt")

	// Create a reader that will fail after writing some data
	testData := []byte("This is test data that should not be partially written")
	reader := &errorReader{
		data:      testData,
		failAfter: 10, // Fail after 10 bytes
	}

	// Attempt to write - this should fail
	err := Write(targetFile, reader)
	if err == nil {
		t.Fatal("Expected Write to fail, but it succeeded")
	}

	// Verify that NO partial file was left behind
	// This is the correct behavior - atomic write should clean up on failure
	if _, err := os.Stat(targetFile); err == nil {
		t.Errorf("Partial file should not exist at %s after failed write", targetFile)
	} else if !os.IsNotExist(err) {
		t.Fatalf("Unexpected error checking for file: %v", err)
	}
}
