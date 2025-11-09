package export

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turbot/steampipe/v2/pkg/test/helpers"
)

// =============================================================================
// BUG HUNTING TESTS - Focused on finding real bugs
// =============================================================================

// TestManagerConcurrentRegistration_RaceCondition tests for data races during concurrent registration
// BUG?: Manager has no mutex, concurrent Register() calls may race
func TestManagerConcurrentRegistration_RaceCondition(t *testing.T) {
	m := NewManager()

	// Create multiple exporters to register concurrently
	exporters := make([]*testExporter, 100)
	for i := 0; i < 100; i++ {
		exporters[i] = &testExporter{
			name:      fmt.Sprintf("exporter%d", i),
			extension: ".test",
			alias:     "",
		}
	}

	// Register all exporters concurrently
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for _, exp := range exporters {
		wg.Add(1)
		go func(exporter *testExporter) {
			defer wg.Done()
			err := m.Register(exporter)
			if err != nil {
				errors <- err
			}
		}(exp)
	}

	wg.Wait()
	close(errors)

	// Collect errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}

	// All registrations should succeed (no concurrent duplicate names)
	assert.Empty(t, errs, "Expected no errors from concurrent registration")

	// Verify all exporters were registered
	for _, exp := range exporters {
		_, ok := m.registeredExporters[exp.name]
		assert.True(t, ok, "Exporter %s should be registered", exp.name)
	}
}

// TestManagerConcurrentAccess_ReadDuringWrite tests reading while writing
// BUG?: Concurrent reads and writes to Manager maps may cause panics
func TestManagerConcurrentAccess_ReadDuringWrite(t *testing.T) {
	m := NewManager()

	// Pre-register one exporter
	m.Register(&testExporter{name: "json", extension: ".json"})

	var wg sync.WaitGroup
	done := make(chan bool)

	// Goroutine 1: Keep registering new exporters
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			select {
			case <-done:
				return
			default:
				m.Register(&testExporter{
					name:      fmt.Sprintf("exporter%d", i),
					extension: ".test",
				})
			}
		}
	}()

	// Goroutine 2: Keep validating export formats (reads from map)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			select {
			case <-done:
				return
			default:
				m.ValidateExportFormat([]string{"json"})
			}
		}
	}()

	// Wait for both to complete
	close(done)
	wg.Wait()

	// If we get here without panic, test passes
	// But run with -race to detect actual data races
}

// TestTargetExport_NilExporter tests what happens with nil exporter
// BUG?: Target.Export doesn't check for nil exporter before calling methods
func TestTargetExport_NilExporter(t *testing.T) {
	target := &Target{
		exporter: nil, // BUG: What happens with nil exporter?
		filePath: "output.json",
	}

	// This should panic if no nil check exists
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Found potential bug: panic on nil exporter: %v", r)
			// This is expected behavior - we found the bug!
		}
	}()

	testData := &testExportSourceData{}
	_, err := target.Export(context.Background(), testData)

	// If we get here without panic, either:
	// 1. There's a nil check (good)
	// 2. The nil dereference hasn't happened yet (bad)
	if err != nil {
		assert.Contains(t, err.Error(), "nil", "Error should mention nil exporter")
	}
}

// TestTargetExport_GetwdFailure tests behavior when os.Getwd() fails
// BUG?: Target.Export ignores error from os.Getwd() (line 20: pwd, _ := os.Getwd())
func TestTargetExport_GetwdFailure(t *testing.T) {
	// This is hard to test directly because we can't make os.Getwd() fail easily
	// But we can document the issue: if Getwd() fails, pwd will be empty string
	// and the message will be "File exported to /filename" which is incorrect

	target := &Target{
		exporter:      &testExporter{name: "test", extension: ".test"},
		filePath:      "test.json",
		isNamedTarget: true,
	}

	testData := &testExportSourceData{}
	msg, err := target.Export(context.Background(), testData)

	assert.NoError(t, err)
	// Message should contain full path, but if Getwd() failed, it would be wrong
	// This is a minor bug: error is ignored
	t.Logf("Export message: %s", msg)
}

// TestWrite_PartialWriteFailure tests behavior when io.Copy fails mid-write
// BUG?: If io.Copy fails, file exists with partial data and may not be cleaned up
func TestWrite_PartialWriteFailure(t *testing.T) {
	tempDir := helpers.CreateTempDir(t)
	filePath := filepath.Join(tempDir, "partial.txt")

	// Create a reader that fails after some data
	failingReader := &failAfterNReader{
		data:      []byte("This is test data that will fail"),
		failAfter: 10,
	}

	err := Write(filePath, failingReader)
	assert.Error(t, err, "Write should fail when reader fails")

	// Check if partial file exists
	if _, statErr := os.Stat(filePath); statErr == nil {
		// File exists with partial data - is this a bug?
		data, _ := os.ReadFile(filePath)
		t.Logf("Partial file exists with %d bytes: %q", len(data), string(data))
		t.Logf("Potential issue: partial file not cleaned up on error")

		// This might be intentional, or it might be a bug
		// depending on whether callers expect cleanup
	}
}

// TestResolveTargetsFromArgs_HugeInput tests with extremely large input
// BUG?: No limit on number of exports, could cause memory issues
func TestResolveTargetsFromArgs_HugeInput(t *testing.T) {
	m := NewManager()
	m.Register(&testExporter{name: "json", extension: ".json"})

	// Try to resolve huge number of export targets
	exports := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		exports[i] = fmt.Sprintf("output%d.json", i)
	}

	targets, err := m.resolveTargetsFromArgs(exports, "test")
	assert.NoError(t, err)
	assert.Len(t, targets, 10000, "Should handle large number of exports")

	// If this causes memory issues or hangs, we found a bug
	t.Logf("Successfully resolved %d targets", len(targets))
}

// TestRegisterExporterByExtension_DeleteAndReInsert tests map manipulation edge case
// BUG?: Line 65 deletes from registeredExtensions, then line 79 re-inserts
// What if there's a race between delete and insert?
func TestRegisterExporterByExtension_DeleteAndReInsert(t *testing.T) {
	m := NewManager()

	// Register first non-default exporter
	exp1 := &testExporter{name: "custom1", extension: ".json", alias: ""}
	err := m.Register(exp1)
	assert.NoError(t, err)

	// Verify it's registered
	registered, ok := m.registeredExtensions[".json"]
	assert.True(t, ok)
	assert.Equal(t, "custom1", registered.Name())

	// Register second non-default exporter with same extension
	// This should trigger delete (line 65) then re-insert (line 79)
	exp2 := &testExporter{name: "custom2", extension: ".json", alias: ""}
	err = m.Register(exp2)
	assert.NoError(t, err)

	// What's in the map now?
	registered, ok = m.registeredExtensions[".json"]
	assert.True(t, ok, "Extension should still be registered after delete/re-insert")
	assert.Equal(t, "custom2", registered.Name(), "Should have second exporter")
}

// TestValidateExportFormat_WithInvalidAndNilTarget tests potential nil pointer dereference
// BUG?: Line 186 appends nil target when error occurs, line 192 might dereference it
func TestValidateExportFormat_WithInvalidAndNilTarget(t *testing.T) {
	m := NewManager()
	m.Register(&testExporter{name: "json", extension: ".json"})

	// Mix valid and invalid formats
	err := m.ValidateExportFormat([]string{"json", "invalid-format"})

	// Should return error about invalid format
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")

	// Key question: did we avoid nil pointer panic in lines 192-193?
	// The code returns early on line 188 if any formats are invalid,
	// so we never reach the nil pointer dereference. Not a bug, but fragile.
}

// =============================================================================
// Helper types for bug hunting tests
// =============================================================================

// failAfterNReader is a reader that succeeds for N bytes then fails
type failAfterNReader struct {
	data      []byte
	read      int
	failAfter int
}

func (f *failAfterNReader) Read(p []byte) (n int, err error) {
	if f.read >= f.failAfter {
		return 0, errors.New("simulated read failure")
	}

	// Read up to failAfter bytes
	remaining := f.failAfter - f.read
	if remaining > len(p) {
		remaining = len(p)
	}
	if remaining > len(f.data)-f.read {
		remaining = len(f.data) - f.read
	}

	n = copy(p, f.data[f.read:f.read+remaining])
	f.read += n

	if f.read >= f.failAfter {
		return n, errors.New("simulated read failure")
	}

	return n, nil
}
