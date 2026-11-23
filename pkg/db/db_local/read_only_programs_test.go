package db_local

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

// makeProgramsFolderReadOnly leaves the database programs where they are, and makes the db folder and the version
// folder read-only, as a container with its programs in a read-only image sees them. The data and internal folders
// stay writable, as the volumes mounted there are.
func makeProgramsFolderReadOnly(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("folder permissions do not make a folder read-only on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root writes to a read-only folder")
	}
	writeFakePrograms(t)
	// create the folders a start writes to before the version folder is closed
	dataLocation := filepaths.GetDataLocation()
	filepaths.EnsureInternalDir()
	filepaths.GetSetupUnfinishedMarkerLocation()

	versionDir := filepaths.DatabaseInstanceDir()
	dbDir := filepaths.EnsureDatabaseDir()
	for _, dir := range []string{versionDir, dbDir} {
		if err := os.Chmod(dir, 0555); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		_ = os.Chmod(dbDir, 0755)
		_ = os.Chmod(versionDir, 0755)
	})
	// the folder really is read-only here
	probe := filepath.Join(versionDir, "probe")
	if err := os.WriteFile(probe, nil, 0600); err == nil {
		_ = os.Remove(probe)
		t.Skip("this filesystem lets a read-only folder be written to")
	}
	if err := os.WriteFile(filepath.Join(dataLocation, "probe"), nil, 0600); err != nil {
		t.Fatalf("the data folder must stay writable: %v", err)
	}
	_ = os.Remove(filepath.Join(dataLocation, "probe"))
	return versionDir
}

func namesIn(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}

func assertVersionFolderUntouched(t *testing.T, versionDir string) {
	t.Helper()
	if got := strings.Join(namesIn(t, versionDir), ","); got != "data,postgres" {
		t.Errorf("a start must write nothing beside the programs and the data folder, found %s", got)
	}
}

// With the programs installed, a start checks the data folder, and prepareDb sets up a database where there is none
// with installDatabase, the same step a first install uses. Nothing else in a start writes under the db folder.
func startWithProgramsInstalled(ctx context.Context) error {
	if err := checkDataFolderUsable(); err != nil {
		return err
	}
	if needsInit() {
		return installDatabase(ctx, nil)
	}
	return nil
}

func TestStart_SetsUpNewDatabaseWithReadOnlyProgramsFolder(t *testing.T) {
	setCleanupInstallDir(t)
	calls := stubInstallStepsCountingCalls(t)
	initNewDatabase = func(context.Context, *string) error {
		calls.init++
		return os.WriteFile(filepaths.GetPgVersionLocation(), []byte(currentMajor()+"\n"), 0600)
	}
	versionDir := makeProgramsFolderReadOnly(t)

	// the marker is written while the setup runs, in a folder that can be written
	var markerDuringSetup bool
	setup := initNewDatabase
	initNewDatabase = func(ctx context.Context, name *string) error {
		_, err := os.Stat(unfinishedMarker())
		markerDuringSetup = err == nil
		return setup(ctx, name)
	}

	if err := startWithProgramsInstalled(context.Background()); err != nil {
		t.Fatalf("a new database must be set up with the programs folder read-only: %v", err)
	}
	if calls.init != 1 {
		t.Errorf("expected one setup, got %d", calls.init)
	}
	if !markerDuringSetup {
		t.Error("the setup must be recorded as unfinished while it runs")
	}
	assertGone(t, unfinishedMarker())
	assertVersionFolderUntouched(t, versionDir)
}

func TestCheckDataFolderUsable_StartsExistingDatabaseWithReadOnlyProgramsFolder(t *testing.T) {
	setCleanupInstallDir(t)
	if err := os.WriteFile(filepaths.GetPgVersionLocation(), []byte(currentMajor()+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	versionDir := makeProgramsFolderReadOnly(t)

	if err := checkDataFolderUsable(); err != nil {
		t.Fatalf("an existing database must start with the programs folder read-only: %v", err)
	}
	assertVersionFolderUntouched(t, versionDir)
}

func TestCheckDataFolderUsable_NoDatabaseAndNoMarkerWithReadOnlyProgramsFolder(t *testing.T) {
	setCleanupInstallDir(t)
	versionDir := makeProgramsFolderReadOnly(t)

	// nothing to remove is not an error, though the read-only folder answers a removal with "read-only file system"
	if err := checkDataFolderUsable(); err != nil {
		t.Fatalf("removing a marker that is not there must not fail: %v", err)
	}
	assertVersionFolderUntouched(t, versionDir)
}

func TestStart_SetupKilledPartwayWithReadOnlyProgramsFolderStopsNextStart(t *testing.T) {
	setCleanupInstallDir(t)
	calls := stubInstallStepsCountingCalls(t)
	// initdb wrote PG_VERSION, and the process was killed before the setup ended: nothing removes the marker
	initNewDatabase = func(context.Context, *string) error {
		calls.init++
		if err := os.WriteFile(filepaths.GetPgVersionLocation(), []byte(currentMajor()+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
		return errors.New("Configuring database... FAILED!")
	}
	versionDir := makeProgramsFolderReadOnly(t)

	if err := startWithProgramsInstalled(context.Background()); err == nil {
		t.Fatal("expected the first setup to fail")
	}
	assertExists(t, unfinishedMarker())

	err := startWithProgramsInstalled(context.Background())
	if err == nil {
		t.Fatal("expected the next start to stop: the setup did not finish")
	}
	for _, want := range []string{"did not finish", filepaths.GetDataLocation(), unfinishedMarker()} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the message should contain %q: %v", want, err)
		}
	}
	if calls.init != 1 {
		t.Errorf("the setup must not run again on the unfinished data folder, ran %d times", calls.init)
	}
	assertExists(t, unfinishedMarker())
	assertVersionFolderUntouched(t, versionDir)
}

func TestCheckDataFolderUsable_MarkerInInternalFolderWithoutDatabaseIsRemoved(t *testing.T) {
	setCleanupInstallDir(t)
	if err := os.WriteFile(unfinishedMarker(), nil, 0600); err != nil {
		t.Fatal(err)
	}
	makeProgramsFolderReadOnly(t)

	if err := checkDataFolderUsable(); err != nil {
		t.Fatal(err)
	}
	assertGone(t, unfinishedMarker())
}

func TestCheckDataFolderUsable_MarkerOfOtherVersionIsLeftAlone(t *testing.T) {
	setCleanupInstallDir(t)
	other := filepaths.SetupUnfinishedMarkerLocationForVersion("13.0.0")
	if err := os.WriteFile(other, nil, 0600); err != nil {
		t.Fatal(err)
	}

	if err := checkDataFolderUsable(); err != nil {
		t.Fatal(err)
	}
	assertExists(t, other)
}

func TestCheckDataFolderUsable_StopsOnMarkerLeftBy251(t *testing.T) {
	setCleanupInstallDir(t)
	if err := os.WriteFile(filepaths.GetPgVersionLocation(), []byte(currentMajor()+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacyUnfinishedMarker(), nil, 0600); err != nil {
		t.Fatal(err)
	}

	err := checkDataFolderUsable()
	if err == nil {
		t.Fatal("expected a stop: a 2.5.1 setup of this data folder did not finish")
	}
	if !strings.Contains(err.Error(), legacyUnfinishedMarker()) {
		t.Errorf("the message should name the marker to delete: %v", err)
	}
	assertExists(t, legacyUnfinishedMarker())
}

func TestCheckDataFolderUsable_RemovesMarkerLeftBy251WhenNoDatabase(t *testing.T) {
	setCleanupInstallDir(t)
	if err := os.WriteFile(legacyUnfinishedMarker(), nil, 0600); err != nil {
		t.Fatal(err)
	}

	if err := checkDataFolderUsable(); err != nil {
		t.Fatal(err)
	}
	assertGone(t, legacyUnfinishedMarker())
}
