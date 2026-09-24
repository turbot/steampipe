package installationstate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

func TestInstallationStateSaveFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	app_specific.InstallDir = filepath.Join(tempDir, ".steampipe")

	state := newInstallationState()

	if err := state.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	path := filepaths.StateFilePath()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat %s: %v", path, err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected %s to have permissions 0600, got %o", path, perm)
	}
}
