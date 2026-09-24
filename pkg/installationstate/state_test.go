package installationstate

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
)

// TestConcurrentSaveLoad demonstrates bug #5038: Save() is not atomic
// (os.Remove followed by os.WriteFile), so a concurrent Load() can hit a
// TOCTOU race between its file-existence check and its read -- the file
// vanishes or is mid-write after Load() sees it exist -- and return an
// error, even though a valid state file was written moments before and
// after.
func TestConcurrentSaveLoad(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "steampipe-installationstate-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	app_specific.InstallDir = filepath.Join(tempDir, ".steampipe")

	// seed an initial valid state file, as would exist on a real install
	seed := newInstallationState()
	if err := seed.Save(); err != nil {
		t.Fatalf("failed to seed initial state: %v", err)
	}

	const concurrency = 20
	const iterations = 200

	var wg sync.WaitGroup
	errCh := make(chan error, concurrency*iterations*2)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				s := newInstallationState()
				if err := s.Save(); err != nil {
					errCh <- fmt.Errorf("save: %w", err)
				}
			}
		}()
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if _, err := Load(); err != nil {
					errCh <- fmt.Errorf("load: %w", err)
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)

	var errs []error
	for e := range errCh {
		errs = append(errs, e)
	}
	if len(errs) > 0 {
		t.Fatalf("got %d error(s) from concurrent Save()/Load() against a file that always had valid content written to it; first: %v", len(errs), errs[0])
	}
}
