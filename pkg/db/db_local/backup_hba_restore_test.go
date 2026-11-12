package db_local

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

// the root-only file an older Steampipe wrote: root connects over TCP, which the upgrade's rules do not
// allow, and which the upgrade replaces
const olderSteampipeHba = "hostssl all root samehost trust\nhost all root samehost trust\n"

func TestStartDatabaseInLocation_StartFailureRestoresOldPgHba(t *testing.T) {
	oldLocation := filepath.Join(shortTempDir(t), "old")
	oldDataDir := filepath.Join(oldLocation, "data")
	if err := os.MkdirAll(oldDataDir, 0700); err != nil {
		t.Fatal(err)
	}
	hbaPath := filepath.Join(oldDataDir, pgHbaFileName)
	if err := os.WriteFile(hbaPath, []byte(olderSteampipeHba), 0600); err != nil {
		t.Fatal(err)
	}

	// there is no postgres binary, so the server cannot start after pg_hba.conf has been rewritten
	if _, err := startDatabaseInLocation(context.Background(), oldLocation); err == nil {
		t.Fatal("expected the start to fail")
	}

	got, err := os.ReadFile(hbaPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != olderSteampipeHba {
		t.Errorf("the old pg_hba.conf must be restored exactly, got:\n%s", got)
	}
}

func TestStartDatabaseInLocation_StartFailureRemovesPgHbaThatDidNotExist(t *testing.T) {
	oldLocation := filepath.Join(shortTempDir(t), "old")
	oldDataDir := filepath.Join(oldLocation, "data")
	if err := os.MkdirAll(oldDataDir, 0700); err != nil {
		t.Fatal(err)
	}

	if _, err := startDatabaseInLocation(context.Background(), oldLocation); err == nil {
		t.Fatal("expected the start to fail")
	}

	if _, err := os.Stat(filepath.Join(oldDataDir, pgHbaFileName)); !os.IsNotExist(err) {
		t.Errorf("a pg_hba.conf that did not exist must not be left behind, stat err: %v", err)
	}
}

func TestMoveDataFromSameMajorInstall_FailureRestoresOldPgHba(t *testing.T) {
	original := app_specific.InstallDir
	defer func() { app_specific.InstallDir = original }()
	app_specific.InstallDir = filepath.Join(shortTempDir(t), ".steampipe")

	oldLocation := filepath.Join(app_specific.InstallDir, "db", "14.19.0")
	writeFakeInstall(t, oldLocation)
	hbaPath := filepath.Join(oldLocation, "data", pgHbaFileName)
	if err := os.WriteFile(hbaPath, []byte(olderSteampipeHba), 0600); err != nil {
		t.Fatal(err)
	}

	// the fake postgres binary cannot be executed, so starting the old server fails
	err := moveDataFromSameMajorInstall(context.Background(), oldLocation, filepaths.GetDataLocation())
	if err == nil {
		t.Fatal("expected an error")
	}

	got, readErr := os.ReadFile(hbaPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != olderSteampipeHba {
		t.Errorf("the old pg_hba.conf must be restored exactly, got:\n%s", got)
	}
}

func TestAbortInstallAfterBackupError_SocketPathTooLongStopsTheInstall(t *testing.T) {
	original := app_specific.InstallDir
	defer func() { app_specific.InstallDir = original }()
	app_specific.InstallDir = filepath.Join(t.TempDir(), strings.Repeat("x", 100), ".steampipe")

	oldDataDir := filepath.Join(app_specific.InstallDir, "db", "14.19.0", "data")
	backupErr := checkSocketPathLengthForUpgrade(oldDataDir, 9193)
	if backupErr == nil {
		t.Fatal("the socket path is over the limit")
	}
	newLocation := filepaths.DatabaseInstanceDir()
	_ = filepaths.GetDatabaseLocation()
	_ = filepaths.GetDataLocation()

	err := abortInstallAfterBackupError(backupErr)
	if err == nil {
		t.Fatal("the install must stop")
	}
	if !strings.Contains(err.Error(), "Move "+app_specific.InstallDir) || !strings.Contains(err.Error(), constants.EnvInstallDir) {
		t.Errorf("the error must be the upgrade advice: %v", err)
	}
	if _, statErr := os.Stat(newLocation); !os.IsNotExist(statErr) {
		t.Errorf("new installation should have been removed so the next run installs again, stat err: %v", statErr)
	}
}

func TestAbortInstallAfterBackupError_OtherErrorsCarryOn(t *testing.T) {
	if err := abortInstallAfterBackupError(errors.New("pg_dump failed")); err != nil {
		t.Errorf("a failed backup does not stop the install: %v", err)
	}
}
