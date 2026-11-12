package db_local

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Masterminds/semver/v3"
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

// a failed install must leave every older folder, removable or not, exactly where it was. Each case makes
// EnsureDBInstalled fail at a different step and lists every path under the older folders beforehand.
func TestEnsureDBInstalled_FailureRemovesNothing(t *testing.T) {
	// older folders: two hold no database and would be removed after a successful install, one holds a database
	// and would be kept; none holds programs, so the install does not take the data-move path
	removableOnly := func(t *testing.T, dbDir string) []string {
		empty := filepath.Join(dbDir, "14.16.0")
		emptyData := filepath.Join(dbDir, "14.15.0")
		for _, dir := range []string{empty, filepath.Join(emptyData, "data")} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
		}
		withData := filepath.Join(dbDir, "14.17.0", "data")
		if err := os.MkdirAll(withData, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(withData, "PG_VERSION"), []byte("14\n"), 0600); err != nil {
			t.Fatal(err)
		}
		return []string{empty, emptyData, filepath.Join(dbDir, "14.17.0")}
	}

	tests := []struct {
		name     string
		layout   func(t *testing.T, dbDir string) []string
		download error
		fdw      error
		running  bool
	}{
		{name: "a previous service is running", layout: removableOnly, running: true},
		{name: "the database download fails", layout: removableOnly, download: errors.New("download failed")},
		{name: "the FDW install fails", layout: removableOnly, fdw: errors.New("fdw failed")},
		// the stubbed downloads leave no initdb binary, so initialising the new database fails
		{name: "initialising the database fails", layout: removableOnly},
		{
			// the folders are the target's major version, so the install moves the data. The data folder is the
			// only one holding both programs and a database, so it is the one moved, and its fake server cannot
			// start, so the move fails
			name: "moving the data folder fails",
			layout: func(t *testing.T, dbDir string) []string {
				target, err := semver.NewVersion(constants.DatabaseVersion)
				if err != nil {
					t.Fatal(err)
				}
				withData := filepath.Join(dbDir, fmt.Sprintf("%d.0.1", target.Major()))
				writeFakeInstall(t, withData)
				programsOnly := filepath.Join(dbDir, fmt.Sprintf("%d.0.2", target.Major()))
				writeProgramsOnlyInstall(t, programsOnly)
				empty := filepath.Join(dbDir, fmt.Sprintf("%d.0.0", target.Major()))
				if err := os.MkdirAll(empty, 0755); err != nil {
					t.Fatal(err)
				}
				return []string{withData, programsOnly, empty}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbDir := setCleanupInstallDir(t)
			// a failed initdb leaves LC_ALL set
			t.Setenv("LC_ALL", "")

			origDownload, origFdw := installDbFiles, installFdwFiles
			t.Cleanup(func() { installDbFiles, installFdwFiles = origDownload, origFdw })
			installDbFiles = func(context.Context) error { return tt.download }
			installFdwFiles = func(context.Context, bool) (string, error) { return "", tt.fdw }

			folders := tt.layout(t, dbDir)
			var before []string
			for _, folder := range folders {
				err := filepath.WalkDir(folder, func(path string, _ fs.DirEntry, err error) error {
					before = append(before, path)
					return err
				})
				if err != nil {
					t.Fatal(err)
				}
			}

			if tt.running {
				// this test process stands in for the running service
				info := newRunningDBInstanceInfo(&exec.Cmd{Process: &os.Process{Pid: os.Getpid()}}, []string{"localhost"}, 9193, "steampipe", "password", constants.InvokerService)
				if err := info.Save(); err != nil {
					t.Fatal(err)
				}
			}

			if err := EnsureDBInstalled(context.Background()); err == nil {
				t.Fatal("expected an error")
			}

			for _, path := range before {
				assertExists(t, path)
			}
		})
	}
}

// a successful install, with the download, FDW install and database initialisation replaced, removes the older
// folders that hold no database and keeps the one that does. No older folder holds both programs and a database,
// so the backup finds nothing to dump and the data move, which starts postgres and cannot be replaced, is not reached.
func TestEnsureDBInstalled_SuccessRemovesFoldersWithoutDatabase(t *testing.T) {
	dbDir := setCleanupInstallDir(t)

	empty := filepath.Join(dbDir, "14.16.0")
	emptyData := filepath.Join(dbDir, "14.15.0")
	for _, dir := range []string{empty, filepath.Join(emptyData, "data")} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	// programs but no database: the install does not pick it to carry over, so this cleanup must remove it
	programsOnly := filepath.Join(dbDir, "14.18.0")
	writeProgramsOnlyInstall(t, programsOnly)
	withData := filepath.Join(dbDir, "14.17.0", "data")
	if err := os.MkdirAll(withData, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(withData, "PG_VERSION"), []byte("14\n"), 0600); err != nil {
		t.Fatal(err)
	}

	origDownload, origFdw, origInit := installDbFiles, installFdwFiles, initNewDatabase
	t.Cleanup(func() { installDbFiles, installFdwFiles, initNewDatabase = origDownload, origFdw, origInit })
	installDbFiles = func(context.Context) error { return nil }
	installFdwFiles = func(context.Context, bool) (string, error) { return "", nil }
	initNewDatabase = func(context.Context, *string) error { return nil }

	if err := EnsureDBInstalled(context.Background()); err != nil {
		t.Fatal(err)
	}

	assertGone(t, empty)
	assertGone(t, emptyData)
	assertGone(t, programsOnly)
	assertExists(t, filepath.Join(withData, "PG_VERSION"))
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
