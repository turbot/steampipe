package db_local

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

func TestRunningDBInstanceInfoSaveFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	app_specific.InstallDir = filepath.Join(tempDir, ".steampipe")

	info := newRunningDBInstanceInfo(&exec.Cmd{Process: &os.Process{Pid: 1234}}, []string{"localhost"}, 9193, "steampipe", "password", constants.InvokerService)

	if err := info.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	path := filepaths.RunningInfoFilePath()
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat %s: %v", path, err)
	}
	if perm := stat.Mode().Perm(); perm != 0600 {
		t.Errorf("expected %s to have permissions 0600, got %o", path, perm)
	}
}

// Versions up to v2.4.9 created the state file as 0644. Saving over such a file
// must narrow it too, since it holds the database password.
func TestRunningDBInstanceInfoSaveNarrowsExistingFilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	originalInstallDir := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = originalInstallDir })
	app_specific.InstallDir = filepath.Join(tempDir, ".steampipe")

	path := filepaths.RunningInfoFilePath()
	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil { //nolint:gosec // G306: recreates the 0644 file that older versions wrote
		t.Fatalf("failed to write %s: %v", path, err)
	}

	info := newRunningDBInstanceInfo(&exec.Cmd{Process: &os.Process{Pid: 1234}}, []string{"localhost"}, 9193, "steampipe", "password", constants.InvokerService)
	if err := info.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	stat, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat %s: %v", path, err)
	}
	if perm := stat.Mode().Perm(); perm != 0600 {
		t.Errorf("expected %s to have permissions 0600, got %o", path, perm)
	}
}
