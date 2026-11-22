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
// allow, so the upgrade must never replace it
const olderSteampipeHba = "hostssl all root samehost trust\nhost all root samehost trust\n"

func TestStartDatabaseInLocation_StartFailureLeavesOldPgHbaAlone(t *testing.T) {
	installDir := shortTempDir(t)
	original := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = original })
	app_specific.InstallDir = filepath.Join(installDir, ".steampipe")
	oldLocation := filepath.Join(installDir, "old")
	oldDataDir := filepath.Join(oldLocation, "data")
	if err := os.MkdirAll(oldDataDir, 0700); err != nil {
		t.Fatal(err)
	}
	hbaPath := filepath.Join(oldDataDir, pgHbaFileName)
	if err := os.WriteFile(hbaPath, []byte(olderSteampipeHba), 0600); err != nil {
		t.Fatal(err)
	}

	// there is no postgres binary, so the server cannot start
	if _, err := startDatabaseInLocation(context.Background(), oldLocation); err == nil {
		t.Fatal("expected the start to fail")
	}

	got, err := os.ReadFile(hbaPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != olderSteampipeHba {
		t.Errorf("the old pg_hba.conf must not change, got:\n%s", got)
	}
}

func TestStartDatabaseInLocation_StartFailureRemovesPgHbaThatDidNotExist(t *testing.T) {
	installDir := shortTempDir(t)
	original := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = original })
	app_specific.InstallDir = filepath.Join(installDir, ".steampipe")
	oldLocation := filepath.Join(installDir, "old")
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

func TestMoveDataFromSameMajorInstall_FailureLeavesOldPgHbaAlone(t *testing.T) {
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
		t.Errorf("the old pg_hba.conf must not change, got:\n%s", got)
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

func TestPgHbaConfTarget_FollowsAMissingFileLinkUpFromALinkedFolder(t *testing.T) {
	root := t.TempDir()
	real := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(real, 0700); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "data")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, filepath.Join(dir, "sub")); err != nil {
		t.Fatal(err)
	}
	// the OS goes up from a/b, the folder that sub links to, to a/missing_hba.conf
	if err := os.Symlink(filepath.Join("..", "missing_hba.conf"), filepath.Join(real, "middle")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join("sub", "middle"), filepath.Join(dir, pgHbaFileName)); err != nil {
		t.Fatal(err)
	}

	got, isLink, err := pgHbaConfTarget(dir)
	if err != nil {
		t.Fatal(err)
	}
	resolvedA, err := filepath.EvalSymlinks(filepath.Join(root, "a"))
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(resolvedA, "missing_hba.conf"); got != want || !isLink {
		t.Errorf("want the link's end %s, got %s (link %v)", want, got, isLink)
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

func assertOldPgHbaUnchanged(t *testing.T, hbaPath string) {
	t.Helper()
	got, err := os.ReadFile(hbaPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != olderSteampipeHba {
		t.Errorf("the old pg_hba.conf must not change, got:\n%s", got)
	}
}

func TestMoveDataFromSameMajorInstall_StopFailureLeavesOldPgHbaAlone(t *testing.T) {
	oldLocation, hbaPath := fakeOldInstall(t, "14.19.0", true)

	err := moveDataFromSameMajorInstall(context.Background(), oldLocation, filepaths.GetDataLocation())
	if err == nil || !strings.Contains(err.Error(), "could not stop the database") {
		t.Fatalf("expected the stop to fail, got: %v", err)
	}
	assertOldPgHbaUnchanged(t, hbaPath)
}

func TestMoveDataFromSameMajorInstall_MoveFailureLeavesOldPgHbaAlone(t *testing.T) {
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
	assertOldPgHbaUnchanged(t, hbaPath)
}

func TestPrepareBackup_LeavesOldPgHbaAloneWhenTheBackupFails(t *testing.T) {
	_, hbaPath := fakeOldInstall(t, "13.0.0", false)
	tmp := shortTempDir(t)
	t.Setenv("TMPDIR", tmp)

	// there is no pg_dump binary, so the backup fails once the old server is running
	name, err := prepareBackup(context.Background())
	if err == nil {
		t.Fatal("expected the backup to fail")
	}
	if name == nil || *name != fakeDatabaseName {
		t.Fatalf("the old server was running, so its database name is expected, got: %v", name)
	}
	assertOldPgHbaUnchanged(t, hbaPath)
	if entries, readErr := os.ReadDir(tmp); readErr != nil || len(entries) != 0 {
		t.Errorf("no rules file may be left behind, got %v (err %v)", entries, readErr)
	}
}

func TestAbortInstallAfterBackupError_SocketPathTooLongStopsTheInstall(t *testing.T) {
	original := app_specific.InstallDir
	defer func() { app_specific.InstallDir = original }()
	app_specific.InstallDir = filepath.Join(t.TempDir(), strings.Repeat("x", 100), ".steampipe")

	oldDataDir := filepath.Join(app_specific.InstallDir, "db", "14.19.0", "data")
	backupErr := checkSocketPathLength(oldDataDir, 9193)
	if backupErr == nil {
		t.Fatal("test setup: the socket path must be over the limit")
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
	backupErr := checkSocketPathLength(oldDataDir, 9193)

	err := abortInstallAfterBackupError(backupErr)
	if !errors.Is(err, errSocketPathTooLong) {
		t.Fatalf("the install must stop with the socket path error: %v", err)
	}
	if !strings.Contains(err.Error(), newDataLocation+" holds data and was kept") {
		t.Errorf("the error must say the new data folder was kept: %v", err)
	}
	// moving the old installation's folder would leave its data behind
	if !strings.Contains(err.Error(), "Once the install directory is moved") {
		t.Errorf("the error must say the next run upgrades once the install directory is moved: %v", err)
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

// startOldServer starts the fake old server of an installation through startDatabaseInLocation and returns
// what it was started with. It never cleans up the rules file: a Steampipe killed part way does not either, so
// the file goes in a temporary folder of the test's own.
func startOldServer(t *testing.T, oldLocation string) (running *pgRunningInfo, args, hbaFileContent string) {
	t.Helper()
	recordDir := shortTempDir(t)
	t.Setenv(fakePostgresRecordDir, recordDir)
	t.Setenv("TMPDIR", shortTempDir(t))
	running, err := startDatabaseInLocation(context.Background(), oldLocation)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = running.stop(context.Background()) })
	argsBytes, err := os.ReadFile(filepath.Join(recordDir, "args"))
	if err != nil {
		t.Fatal(err)
	}
	hbaBytes, _ := os.ReadFile(filepath.Join(recordDir, "hba"))
	return running, string(argsBytes), string(hbaBytes)
}

// the kill in #5148: the old data directory must already be as the older version left it while the upgrade's
// server runs, so a Steampipe that never gets to clean up leaves nothing behind
func TestStartDatabaseInLocation_KilledUpgradeLeavesTheOldPgHbaAsItWas(t *testing.T) {
	oldLocation, hbaPath := fakeOldInstall(t, "13.0.0", false)
	oldDataDir := filepath.Join(oldLocation, "data")
	past := time.Now().Add(-time.Hour).Truncate(time.Second)
	if err := os.Chtimes(hbaPath, past, past); err != nil {
		t.Fatal(err)
	}

	_, args, hba := startOldServer(t, oldLocation)

	assertOldPgHbaUnchanged(t, hbaPath)
	if info, err := os.Stat(hbaPath); err != nil || !info.ModTime().Equal(past) {
		t.Errorf("the old pg_hba.conf must not be written at all, stat: %v, %v", info, err)
	}
	if _, err := os.Stat(filepath.Join(oldDataDir, pgHbaDirName)); !os.IsNotExist(err) {
		t.Errorf("%s must not be made in the old data directory, stat err: %v", pgHbaDirName, err)
	}
	if !strings.Contains(args, "hba_file=") {
		t.Fatalf("the server must be given Steampipe's rules through hba_file, got args:\n%s", args)
	}
	if !strings.HasPrefix(hba, constants.PgHbaHeader) || !strings.HasSuffix(hba, constants.PgHbaSteampipeRules) {
		t.Errorf("the rules the server read are not Steampipe's:\n%s", hba)
	}
}

func TestStartDatabaseInLocation_RulesFileHoldsThePgHbaConfDRules(t *testing.T) {
	oldLocation, hbaPath := fakeOldInstall(t, "13.0.0", false)
	writeFile(t, filepath.Join(oldLocation, "data", pgHbaDirName, "10-mine.conf"), "host all bob all md5\n")

	_, _, hba := startOldServer(t, oldLocation)

	if !strings.Contains(hba, "host all bob all md5") {
		t.Errorf("the user's rules in %s must be in the rules the server read:\n%s", pgHbaDirName, hba)
	}
	assertOldPgHbaUnchanged(t, hbaPath)
}

func TestStartDatabaseInLocation_NoPgHbaConfIsNotCreated(t *testing.T) {
	oldLocation, hbaPath := fakeOldInstall(t, "13.0.0", false)
	if err := os.Remove(hbaPath); err != nil {
		t.Fatal(err)
	}

	_, args, _ := startOldServer(t, oldLocation)

	// a missing pg_hba.conf looks like "not installed" to Steampipe 2.4 and earlier
	if _, err := os.Stat(hbaPath); !os.IsNotExist(err) {
		t.Errorf("a pg_hba.conf that did not exist must not be made, stat err: %v", err)
	}
	if !strings.Contains(args, "hba_file=") {
		t.Errorf("the server must be given Steampipe's rules through hba_file, got args:\n%s", args)
	}
}

func TestStartDatabaseInLocation_UsersOwnPgHbaConfIsUsedWhereItIs(t *testing.T) {
	oldLocation, hbaPath := fakeOldInstall(t, "13.0.0", false)
	userHba := "local all all trust\nhost all bob all md5\n"
	writeFile(t, hbaPath, userHba)

	_, args, _ := startOldServer(t, oldLocation)

	if strings.Contains(args, "hba_file=") {
		t.Errorf("the user's own file must be read where it is, got args:\n%s", args)
	}
	if got := readHba(t, filepath.Join(oldLocation, "data")); got != userHba {
		t.Errorf("the user's file changed:\n%s", got)
	}
}

func TestStartDatabaseInLocation_RulesFileIsInTheTempFolderNotTheDataFolder(t *testing.T) {
	oldLocation, _ := fakeOldInstall(t, "13.0.0", false)
	tmp := shortTempDir(t)
	t.Setenv("TMPDIR", tmp)

	_, args, _ := startOldServer(t, oldLocation)
	hbaFile := ""
	for _, line := range strings.Split(args, "\n") {
		if f, ok := strings.CutPrefix(line, "hba_file="); ok {
			hbaFile = f
		}
	}
	if hbaFile == "" {
		t.Fatalf("the server was not given an hba_file, args:\n%s", args)
	}
	if strings.HasPrefix(hbaFile, oldLocation) {
		t.Errorf("the rules file %s must be outside the old installation", hbaFile)
	}
	if _, err := os.Stat(hbaFile); err != nil {
		t.Errorf("the rules file must exist while the server runs: %v", err)
	}
}

func TestMoveDataFromSameMajorInstall_SuccessLeavesNoRulesFile(t *testing.T) {
	oldLocation, _ := fakeOldInstall(t, "14.19.0", false)
	tmp := shortTempDir(t)
	t.Setenv("TMPDIR", tmp)

	if err := moveDataFromSameMajorInstall(context.Background(), oldLocation, filepaths.GetDataLocation()); err != nil {
		t.Fatal(err)
	}

	if entries, err := os.ReadDir(tmp); err != nil || len(entries) != 0 {
		t.Errorf("no rules file may be left behind, got %v (err %v)", entries, err)
	}
	// the data folder moved with the older version's pg_hba.conf, which the next start replaces
	assertOldPgHbaUnchanged(t, filepath.Join(filepaths.GetDataLocation(), pgHbaFileName))
}

func TestStartDatabaseInLocation_StartFailureLeavesNoRulesFile(t *testing.T) {
	installDir := shortTempDir(t)
	original := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = original })
	app_specific.InstallDir = filepath.Join(installDir, ".steampipe")
	oldLocation := filepath.Join(installDir, "old")
	if err := os.MkdirAll(filepath.Join(oldLocation, "data"), 0700); err != nil {
		t.Fatal(err)
	}
	tmp := shortTempDir(t)
	t.Setenv("TMPDIR", tmp)

	// there is no postgres binary, so the server cannot start
	if _, err := startDatabaseInLocation(context.Background(), oldLocation); err == nil {
		t.Fatal("expected the start to fail")
	}

	if entries, err := os.ReadDir(tmp); err != nil || len(entries) != 0 {
		t.Errorf("no rules file may be left behind, got %v (err %v)", entries, err)
	}
}
