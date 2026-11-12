package db_local

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestMoveDataFromSameMajorInstall_EditedPgHbaMessageIsShownAsItIs(t *testing.T) {
	original := app_specific.InstallDir
	defer func() { app_specific.InstallDir = original }()
	app_specific.InstallDir = filepath.Join(shortTempDir(t), ".steampipe")

	oldLocation := filepath.Join(app_specific.InstallDir, "db", "14.19.0")
	writeFakeInstall(t, oldLocation)
	oldDataDir := filepath.Join(oldLocation, "data")
	steampipeFileWith(t, oldDataDir, nil)
	edited := readHba(t, oldDataDir) + "host all bob all md5\n"
	writeFile(t, filepath.Join(oldDataDir, pgHbaFileName), edited)
	newBinaries := filepath.Join(filepaths.DatabaseInstanceDir(), "postgres")
	if err := os.MkdirAll(newBinaries, 0700); err != nil {
		t.Fatal(err)
	}

	err := moveDataFromSameMajorInstall(context.Background(), oldLocation, filepaths.GetDataLocation())
	var stop *pgHbaStopError
	if !errors.As(err, &stop) {
		t.Fatalf("want a pgHbaStopError, got %v", err)
	}
	if !strings.HasPrefix(err.Error(), stop.message) {
		t.Errorf("the message must start with the instructions, not a wrapper:\n%s", err)
	}
	if strings.Contains(err.Error(), "could not move the database") || strings.Contains(err.Error(), "could not start") {
		t.Errorf("the message must not be wrapped:\n%s", err)
	}
	if _, statErr := os.Stat(newBinaries); !os.IsNotExist(statErr) {
		t.Errorf("the new binaries must be removed so the next run retries the move, stat err: %v", statErr)
	}
	if got := readHba(t, oldDataDir); got != edited {
		t.Errorf("the edited file changed:\n%s", got)
	}
}

func TestSavePgHbaConf_RestoreLeavesAnUnchangedFileAlone(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, pgHbaFileName)
	if err := os.WriteFile(path, []byte(olderSteampipeHba), 0600); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-time.Hour).Truncate(time.Second)
	if err := os.Chtimes(path, past, past); err != nil {
		t.Fatal(err)
	}

	restore, err := savePgHbaConf(dir)
	if err != nil {
		t.Fatal(err)
	}
	restore()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(past) {
		t.Errorf("an unchanged pg_hba.conf must not be rewritten, modified at %s", info.ModTime())
	}
}

func TestSavePgHbaConf_RestoreReplacesAChangedFileKeepingItsMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, pgHbaFileName)
	if err := os.WriteFile(path, []byte(olderSteampipeHba), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0640); err != nil {
		t.Fatal(err)
	}

	restore, err := savePgHbaConf(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("host all all all trust\n"), 0600); err != nil {
		t.Fatal(err)
	}
	restore()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != olderSteampipeHba {
		t.Errorf("the changed pg_hba.conf must be put back, got:\n%s", got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0640 {
		t.Errorf("the original mode must be kept, got %v", info.Mode().Perm())
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("no temporary file may be left behind, found %d entries", len(entries))
	}
}

// fakeOldInstall sets up an installation that an older Steampipe wrote, with a postgres binary that starts,
// and returns its location and the path of its pg_hba.conf.
func fakeOldInstall(t *testing.T, version string, stopFails bool) (oldLocation, hbaPath string) {
	t.Helper()
	original := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = original })
	app_specific.InstallDir = filepath.Join(shortTempDir(t), ".steampipe")

	oldLocation = filepath.Join(app_specific.InstallDir, "db", version)
	writeFakeInstall(t, oldLocation)
	writeFakePostgres(t, oldLocation, stopFails)
	hbaPath = filepath.Join(oldLocation, "data", pgHbaFileName)
	if err := os.WriteFile(hbaPath, []byte(olderSteampipeHba), 0600); err != nil {
		t.Fatal(err)
	}
	return oldLocation, hbaPath
}

func assertOldPgHbaRestored(t *testing.T, hbaPath string) {
	t.Helper()
	got, err := os.ReadFile(hbaPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != olderSteampipeHba {
		t.Errorf("the old pg_hba.conf must be restored exactly, got:\n%s", got)
	}
}

func TestMoveDataFromSameMajorInstall_StopFailureRestoresOldPgHba(t *testing.T) {
	oldLocation, hbaPath := fakeOldInstall(t, "14.19.0", true)

	err := moveDataFromSameMajorInstall(context.Background(), oldLocation, filepaths.GetDataLocation())
	if err == nil || !strings.Contains(err.Error(), "could not stop the database") {
		t.Fatalf("expected the stop to fail, got: %v", err)
	}
	assertOldPgHbaRestored(t, hbaPath)
}

func TestMoveDataFromSameMajorInstall_MoveFailureRestoresOldPgHba(t *testing.T) {
	oldLocation, hbaPath := fakeOldInstall(t, "14.19.0", false)

	// a new data folder that holds data cannot be replaced, so the move fails after the server has stopped
	newDataLocation := filepaths.GetDataLocation()
	if err := os.WriteFile(filepath.Join(newDataLocation, "PG_VERSION"), []byte("14\n"), 0600); err != nil {
		t.Fatal(err)
	}

	err := moveDataFromSameMajorInstall(context.Background(), oldLocation, newDataLocation)
	if err == nil || !strings.Contains(err.Error(), "could not prepare") {
		t.Fatalf("expected the move to fail, got: %v", err)
	}
	assertOldPgHbaRestored(t, hbaPath)
}

func TestPrepareBackup_RestoresOldPgHbaWhenTheBackupFails(t *testing.T) {
	_, hbaPath := fakeOldInstall(t, "13.0.0", false)

	// there is no pg_dump binary, so the backup fails once the old server is running
	name, err := prepareBackup(context.Background())
	if err == nil {
		t.Fatal("expected the backup to fail")
	}
	if name == nil || *name != fakeDatabaseName {
		t.Fatalf("the old server was running, so its database name is expected, got: %v", name)
	}
	assertOldPgHbaRestored(t, hbaPath)
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

func TestAbortInstallAfterBackupError_SocketPathTooLongKeepsNewDataThatHoldsData(t *testing.T) {
	original := app_specific.InstallDir
	defer func() { app_specific.InstallDir = original }()
	app_specific.InstallDir = filepath.Join(t.TempDir(), strings.Repeat("x", 100), ".steampipe")

	newDataLocation := filepaths.GetDataLocation()
	if err := os.WriteFile(filepath.Join(newDataLocation, "PG_VERSION"), []byte("14\n"), 0600); err != nil {
		t.Fatal(err)
	}
	oldDataDir := filepath.Join(app_specific.InstallDir, "db", "14.19.0", "data")
	backupErr := checkSocketPathLengthForUpgrade(oldDataDir, 9193)

	err := abortInstallAfterBackupError(backupErr)
	if !errors.Is(err, errSocketPathTooLong) {
		t.Fatalf("the install must stop with the socket path error: %v", err)
	}
	if !strings.Contains(err.Error(), newDataLocation+" holds data and was kept") {
		t.Errorf("the error must say the new data folder was kept: %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(newDataLocation, "PG_VERSION")); statErr != nil {
		t.Errorf("new data should have been kept: %v", statErr)
	}
}

func TestAbortInstallAfterBackupError_PgHbaStopErrorStopsTheInstall(t *testing.T) {
	original := app_specific.InstallDir
	defer func() { app_specific.InstallDir = original }()
	app_specific.InstallDir = filepath.Join(t.TempDir(), ".steampipe")

	newLocation := filepaths.DatabaseInstanceDir()
	_ = filepaths.GetDatabaseLocation()
	_ = filepaths.GetDataLocation()

	backupErr := fmt.Errorf("start: %w", &pgHbaStopError{message: "fix pg_hba.conf"})
	err := abortInstallAfterBackupError(backupErr)
	var stop *pgHbaStopError
	if !errors.As(err, &stop) {
		t.Fatalf("the install must stop with the pg_hba.conf error: %v", err)
	}
	if _, statErr := os.Stat(newLocation); !os.IsNotExist(statErr) {
		t.Errorf("new installation should have been removed so the next run installs again, stat err: %v", statErr)
	}
}

func TestAbortInstallAfterBackupError_PgHbaStopErrorKeepsNewDataThatHoldsData(t *testing.T) {
	original := app_specific.InstallDir
	defer func() { app_specific.InstallDir = original }()
	app_specific.InstallDir = filepath.Join(t.TempDir(), ".steampipe")

	newDataLocation := filepaths.GetDataLocation()
	if err := os.WriteFile(filepath.Join(newDataLocation, "PG_VERSION"), []byte("14\n"), 0600); err != nil {
		t.Fatal(err)
	}

	err := abortInstallAfterBackupError(&pgHbaStopError{message: "fix pg_hba.conf"})
	if err == nil || !strings.Contains(err.Error(), newDataLocation+" holds data and was kept") {
		t.Errorf("the error must say the new data folder was kept: %v", err)
	}
}

func TestAbortInstallAfterBackupError_OtherErrorsCarryOn(t *testing.T) {
	if err := abortInstallAfterBackupError(errors.New("pg_dump failed")); err != nil {
		t.Errorf("an error with no upgrade advice does not stop the install: %v", err)
	}
}
