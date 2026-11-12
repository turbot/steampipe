package db_local

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

func setCleanupInstallDir(t *testing.T) string {
	t.Helper()
	original := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = original })
	app_specific.InstallDir = filepath.Join(t.TempDir(), ".steampipe")
	return filepaths.EnsureDatabaseDir()
}

func writeProgramsOnlyInstall(t *testing.T, location string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(location, "postgres", "bin"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(location, "postgres", "bin", "postgres"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); err != nil {
		t.Errorf("%s should have been kept: %v", path, err)
	}
}

func assertGone(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Errorf("%s should have been removed, stat err: %v", path, err)
	}
}

func TestRemoveOldInstallsWithoutDatabase_RemovesFoldersHoldingNoDatabase(t *testing.T) {
	dbDir := setCleanupInstallDir(t)

	// programs only: the data folder was moved out
	programsOnly := filepath.Join(dbDir, "14.19.0")
	writeProgramsOnlyInstall(t, programsOnly)
	// programs and an empty data folder
	emptyData := filepath.Join(dbDir, "14.17.0")
	writeProgramsOnlyInstall(t, emptyData)
	if err := os.MkdirAll(filepath.Join(emptyData, "data"), 0700); err != nil {
		t.Fatal(err)
	}
	// left by a failed upgrade: an empty folder
	empty := filepath.Join(dbDir, "13.0.0")
	if err := os.MkdirAll(empty, 0755); err != nil {
		t.Fatal(err)
	}

	removeOldInstallsWithoutDatabase()

	assertGone(t, programsOnly)
	assertGone(t, emptyData)
	assertGone(t, empty)
}

func TestRemoveOldInstallsWithoutDatabase_NeverRemovesFolderHoldingDatabase(t *testing.T) {
	dbDir := setCleanupInstallDir(t)

	withBinaries := filepath.Join(dbDir, "14.19.0")
	writeFakeInstall(t, withBinaries)

	// data without programs is still a database
	dataOnly := filepath.Join(dbDir, "14.17.0")
	if err := os.MkdirAll(filepath.Join(dataOnly, "data"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataOnly, "data", "PG_VERSION"), []byte("14\n"), 0600); err != nil {
		t.Fatal(err)
	}

	// a data folder with anything in it is not treated as empty, whatever that is
	partialData := filepath.Join(dbDir, "13.0.0")
	writeProgramsOnlyInstall(t, partialData)
	if err := os.MkdirAll(filepath.Join(partialData, "data", "base"), 0700); err != nil {
		t.Fatal(err)
	}

	// a data entry that is not a folder is never removed
	dataIsFile := filepath.Join(dbDir, "12.0.0")
	writeProgramsOnlyInstall(t, dataIsFile)
	if err := os.WriteFile(filepath.Join(dataIsFile, "data"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}

	removeOldInstallsWithoutDatabase()

	assertExists(t, filepath.Join(withBinaries, "data", "PG_VERSION"))
	assertExists(t, filepath.Join(dataOnly, "data", "PG_VERSION"))
	assertExists(t, filepath.Join(partialData, "data", "base"))
	assertExists(t, filepath.Join(dataIsFile, "data"))
}

func TestRemoveOldInstallsWithoutDatabase_NeverRemovesDataBehindASymlink(t *testing.T) {
	dbDir := setCleanupInstallDir(t)

	external := filepath.Join(t.TempDir(), "external-data")
	if err := os.MkdirAll(external, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(external, "PG_VERSION"), []byte("14\n"), 0600); err != nil {
		t.Fatal(err)
	}
	linked := filepath.Join(dbDir, "14.19.0")
	writeProgramsOnlyInstall(t, linked)
	if err := os.Symlink(external, filepath.Join(linked, "data")); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	removeOldInstallsWithoutDatabase()

	assertExists(t, filepath.Join(linked, "data"))
	assertExists(t, filepath.Join(external, "PG_VERSION"))
}

// a failed install returns before it reaches the cleanup. This drives the one failure that needs no download:
// a running service from a previous version. The failures after the download are not driven here.
func TestEnsureDBInstalled_FailureRemovesNoOldFolder(t *testing.T) {
	dbDir := setCleanupInstallDir(t)

	programsOnly := filepath.Join(dbDir, "14.17.0")
	writeProgramsOnlyInstall(t, programsOnly)

	// this test process stands in for the running service
	running := newRunningDBInstanceInfo(&exec.Cmd{Process: &os.Process{Pid: os.Getpid()}}, []string{"localhost"}, 9193, "steampipe", "password", constants.InvokerService)
	if err := running.Save(); err != nil {
		t.Fatal(err)
	}

	if err := EnsureDBInstalled(context.Background()); err == nil {
		t.Fatal("expected an error")
	}

	assertExists(t, filepath.Join(programsOnly, "postgres", "bin", "postgres"))
}

func TestRemoveOldInstallsWithoutDatabase_KeepsEverythingElse(t *testing.T) {
	dbDir := setCleanupInstallDir(t)

	current := filepath.Join(dbDir, constants.DatabaseVersion)
	writeProgramsOnlyInstall(t, current)
	newer := filepath.Join(dbDir, "99.0.0")
	writeProgramsOnlyInstall(t, newer)
	// a folder that is not a version folder
	other := filepath.Join(dbDir, "data-tank")
	if err := os.MkdirAll(other, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(other, "toc.dat"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}
	// a version folder holding something other than programs and data
	extra := filepath.Join(dbDir, "14.0.0")
	writeProgramsOnlyInstall(t, extra)
	if err := os.WriteFile(filepath.Join(extra, "notes.txt"), []byte(""), 0600); err != nil {
		t.Fatal(err)
	}
	// files directly in the db folder
	dump := filepath.Join(dbDir, "public.dump")
	if err := os.WriteFile(dump, []byte(""), 0600); err != nil {
		t.Fatal(err)
	}

	removeOldInstallsWithoutDatabase()

	assertExists(t, filepath.Join(current, "postgres", "bin", "postgres"))
	assertExists(t, filepath.Join(newer, "postgres", "bin", "postgres"))
	assertExists(t, filepath.Join(other, "toc.dat"))
	assertExists(t, filepath.Join(extra, "notes.txt"))
	assertExists(t, dump)
}
