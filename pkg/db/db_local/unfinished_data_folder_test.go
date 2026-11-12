package db_local

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

func currentMajor() string {
	return strings.Split(constants.DatabaseVersion, ".")[0]
}

func unfinishedMarker() string {
	return filepath.Join(filepaths.DatabaseInstanceDir(), "setup-unfinished")
}

type installStepCalls struct {
	download int
	init     int
}

func stubInstallStepsCountingCalls(t *testing.T) *installStepCalls {
	t.Helper()
	calls := &installStepCalls{}
	origDownload, origFdw, origInit := installDbFiles, installFdwFiles, initNewDatabase
	t.Cleanup(func() { installDbFiles, installFdwFiles, initNewDatabase = origDownload, origFdw, origInit })
	installDbFiles = func(context.Context) error { calls.download++; return nil }
	installFdwFiles = func(context.Context, bool) (string, error) { return "", nil }
	initNewDatabase = func(context.Context, *string) error { calls.init++; return nil }
	return calls
}

func writeFakePrograms(t *testing.T) {
	t.Helper()
	for _, path := range []string{filepaths.GetInitDbBinaryExecutablePath(), filepaths.GetPostgresBinaryExecutablePath()} {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(""), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func removeFakePrograms(t *testing.T) {
	t.Helper()
	if err := os.RemoveAll(filepaths.GetDatabaseLocation()); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureDBInstalled_StopsOnDataFolderWhoseSetupStoppedPartway(t *testing.T) {
	for _, tt := range []struct {
		name         string
		withPrograms bool
	}{
		{"programs missing", false},
		{"programs present", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			setCleanupInstallDir(t)
			calls := stubInstallStepsCountingCalls(t)

			// the first run gets as far as initdb, which writes PG_VERSION, and stops
			initNewDatabase = func(context.Context, *string) error {
				calls.init++
				if err := os.WriteFile(filepaths.GetPgVersionLocation(), []byte(currentMajor()+"\n"), 0600); err != nil {
					t.Fatal(err)
				}
				return errors.New("Configuring database... FAILED!")
			}
			if err := EnsureDBInstalled(context.Background()); err == nil {
				t.Fatal("expected the first install to fail")
			}
			if calls.init != 1 {
				t.Fatalf("expected the first install to run setup once, ran %d times", calls.init)
			}

			if tt.withPrograms {
				writeFakePrograms(t)
			} else {
				removeFakePrograms(t)
			}
			downloadsBefore, initsBefore := calls.download, calls.init
			dataLocation := filepaths.GetDataLocation()

			err := EnsureDBInstalled(context.Background())
			if err == nil {
				t.Fatal("expected an error: the data folder holds a database whose setup did not finish")
			}
			if !strings.Contains(err.Error(), dataLocation) {
				t.Errorf("the message should name the data folder %s: %v", dataLocation, err)
			}
			if calls.init != initsBefore {
				t.Error("the setup must not be run again on the unfinished data folder")
			}
			if calls.download != downloadsBefore {
				t.Error("nothing should be installed over the unfinished data folder")
			}
			assertExists(t, filepaths.GetPgVersionLocation())
		})
	}
}

func TestEnsureDBInstalled_StopsOnDataFolderFromOtherMajorVersion(t *testing.T) {
	for _, tt := range []struct {
		name         string
		withPrograms bool
	}{
		{"programs missing", false},
		{"programs present", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			setCleanupInstallDir(t)
			calls := stubInstallStepsCountingCalls(t)

			dataLocation := filepaths.GetDataLocation()
			versionFile := filepaths.GetPgVersionLocation()
			dataFile := filepath.Join(dataLocation, "base-data")
			if err := os.WriteFile(versionFile, []byte("9\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(dataFile, []byte("old rows"), 0600); err != nil {
				t.Fatal(err)
			}
			if tt.withPrograms {
				writeFakePrograms(t)
			}

			err := EnsureDBInstalled(context.Background())
			if err == nil {
				t.Fatal("expected an error: the data folder is from another Postgres version")
			}
			for _, want := range []string{dataLocation, "Postgres 9", "Postgres " + currentMajor()} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("the message should contain %q: %v", want, err)
				}
			}
			if calls.init != 0 || calls.download != 0 {
				t.Errorf("nothing should be installed or set up: %+v", *calls)
			}

			if content, err := os.ReadFile(versionFile); err != nil || string(content) != "9\n" {
				t.Errorf("PG_VERSION changed: %q, %v", content, err)
			}
			if content, err := os.ReadFile(dataFile); err != nil || string(content) != "old rows" {
				t.Errorf("data file changed: %q, %v", content, err)
			}
			entries, err := os.ReadDir(dataLocation)
			if err != nil || len(entries) != 2 {
				t.Errorf("the data folder should hold exactly the two files it held, got %v, %v", entries, err)
			}
		})
	}
}

func TestEnsureDBInstalled_MarkerWithoutDatabaseDoesNotStopSetup(t *testing.T) {
	setCleanupInstallDir(t)
	calls := stubInstallStepsCountingCalls(t)
	if err := os.WriteFile(unfinishedMarker(), nil, 0600); err != nil {
		t.Fatal(err)
	}

	if err := EnsureDBInstalled(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls.init != 1 {
		t.Errorf("expected the database to be set up once, got %d", calls.init)
	}
	assertGone(t, unfinishedMarker())
}

func TestFindDifferentPgInstallation_SkipsInstallWhoseSetupDidNotFinish(t *testing.T) {
	dbDir := setCleanupInstallDir(t)
	older := filepath.Join(dbDir, "13.0.0")
	writeFakeInstall(t, older)
	if err := os.WriteFile(filepath.Join(older, "setup-unfinished"), nil, 0600); err != nil {
		t.Fatal(err)
	}

	found, location, err := findDifferentPgInstallation(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Errorf("an installation whose setup did not finish must not be carried over, found %s", location)
	}
}

func TestEnsureDBInstalled_ReusesFinishedDataFolder(t *testing.T) {
	t.Run("set up by an older Steampipe", func(t *testing.T) {
		setCleanupInstallDir(t)
		calls := stubInstallStepsCountingCalls(t)
		if err := os.WriteFile(filepaths.GetPgVersionLocation(), []byte(currentMajor()+"\n"), 0600); err != nil {
			t.Fatal(err)
		}

		if err := EnsureDBInstalled(context.Background()); err != nil {
			t.Fatal(err)
		}
		if calls.init != 0 {
			t.Error("a finished data folder must not be set up again")
		}
		assertExists(t, filepaths.GetPgVersionLocation())
	})

	t.Run("set up by this Steampipe", func(t *testing.T) {
		setCleanupInstallDir(t)
		calls := stubInstallStepsCountingCalls(t)
		initNewDatabase = func(context.Context, *string) error {
			calls.init++
			return os.WriteFile(filepaths.GetPgVersionLocation(), []byte(currentMajor()+"\n"), 0600)
		}
		if err := EnsureDBInstalled(context.Background()); err != nil {
			t.Fatal(err)
		}
		if calls.init != 1 {
			t.Fatalf("expected one setup, got %d", calls.init)
		}

		removeFakePrograms(t)
		if err := EnsureDBInstalled(context.Background()); err != nil {
			t.Fatalf("a data folder whose setup finished must be reused: %v", err)
		}
		if calls.init != 1 {
			t.Error("a finished data folder must not be set up again")
		}
		assertExists(t, filepaths.GetPgVersionLocation())
	})
}
