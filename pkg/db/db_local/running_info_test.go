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
