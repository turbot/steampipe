package db_local

import (
	"context"
	"errors"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

func TestPostgresSocketArgs(t *testing.T) {
	got, err := postgresSocketArgs("/data", 9193)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"-c", `unix_socket_directories="/data"`, "-c", "unix_socket_permissions=0700"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPostgresSocketArgs_RefusesTooLongPath(t *testing.T) {
	originalInstallDir := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = originalInstallDir })
	app_specific.InstallDir = "/" + strings.Repeat("d", maxSocketPathLength())
	dataDir := filepath.Join(app_specific.InstallDir, "db", "14.24.0", "data")
	got, err := postgresSocketArgs(dataDir, 9193)
	if err == nil {
		t.Fatalf("a socket path longer than %d bytes must be refused, got %v", maxSocketPathLength(), got)
	}
	// the data directory holds the database: a new install directory would leave it behind
	if !strings.Contains(err.Error(), "Move "+app_specific.InstallDir) || !strings.Contains(err.Error(), constants.EnvInstallDir) {
		t.Errorf("the error must say to move the install directory and point %s at it: %v", constants.EnvInstallDir, err)
	}
	if strings.Contains(err.Error(), "Set "+constants.EnvInstallDir+" to a shorter path") {
		t.Errorf("the error must not say to set %s to a new path: %v", constants.EnvInstallDir, err)
	}
}

func TestCheckSocketPathLength_Limit(t *testing.T) {
	// "/.s.PGSQL.9193" is 14 bytes
	atLimit := "/" + strings.Repeat("d", maxSocketPathLength()-15)
	if got := len(rootSocketPath(atLimit, 9193)); got != maxSocketPathLength() {
		t.Fatalf("test setup: the path is %d bytes, not %d", got, maxSocketPathLength())
	}
	if err := checkSocketPathLength(atLimit, 9193); err != nil {
		t.Errorf("a path of exactly the limit must be accepted: %v", err)
	}
	if err := checkSocketPathLength(atLimit+"d", 9193); err == nil {
		t.Error("a path one byte over the limit must be refused")
	}
}

func TestCreateCmd_HasSocketSettings(t *testing.T) {
	originalInstallDir := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = originalInstallDir })
	// a short install directory: t.TempDir can make the socket path too long
	app_specific.InstallDir = filepath.Join(shortTempDir(t), ".steampipe")
	cmd, err := createCmd(context.Background(), 9193, []string{"127.0.0.1"})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	for _, setting := range []string{`unix_socket_directories="` + filepaths.GetDataLocation() + `"`, "unix_socket_permissions=0700"} {
		if !strings.Contains(args, setting) {
			t.Errorf("service start is missing %s: %s", setting, args)
		}
	}
}

func TestRootConnectionParams_UsesSocketWithoutSSL(t *testing.T) {
	got := rootConnectionParams(9193)
	if _, ok := got["host"]; ok || got["port"] != "9193" || got["user"] != constants.DatabaseSuperUser || got["sslmode"] != "disable" {
		t.Errorf("unexpected params: %v", got)
	}
	if rootSocketPath("/data", 9193) != "/data/.s.PGSQL.9193" {
		t.Errorf("unexpected socket path: %s", rootSocketPath("/data", 9193))
	}
}

func TestQuoteConnStringValue_KeepsPathWithSpacesAndQuotes(t *testing.T) {
	dataDir := `/Users/me/Steampipe dir/it's\here/db/14.24.0/data`
	config, err := pgx.ParseConfig("host=" + quoteConnStringValue(dataDir) + " port=9193 user=root")
	if err != nil {
		t.Fatal(err)
	}
	if config.Host != dataDir {
		t.Errorf("got host %q, want %q", config.Host, dataDir)
	}
}

// writeArgsRecorder replaces the binary at path with a script that writes its arguments, one per line, to
// the returned file.
func writeArgsRecorder(t *testing.T, path string) (argsFile string) {
	t.Helper()
	argsFile = filepath.Join(t.TempDir(), "args")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > '" + argsFile + "'\n"
	if err := os.WriteFile(path, []byte(script), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0700); err != nil {
		t.Fatal(err)
	}
	return argsFile
}

func recordedHost(t *testing.T, argsFile string) string {
	t.Helper()
	b, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	for _, arg := range strings.Split(string(b), "\n") {
		if host, ok := strings.CutPrefix(arg, "--host="); ok {
			return host
		}
	}
	t.Fatalf("no --host in %q", b)
	return ""
}

func TestBackupAndRestoreConnectThroughTheirDataDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the recorders are shell scripts")
	}
	originalInstallDir := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = originalInstallDir })
	app_specific.InstallDir = filepath.Join(t.TempDir(), ".steampipe")
	oldDataDir := filepath.Join(t.TempDir(), "14.19.0", "data")

	dumpArgs := writeArgsRecorder(t, filepaths.PgDumpBinaryExecutablePath())
	if err := takeBackup(context.Background(), &pgRunningInfo{dataDir: oldDataDir, port: 9876, dbName: "steampipe"}); err != nil {
		t.Fatal(err)
	}
	if got := recordedHost(t, dumpArgs); got != oldDataDir {
		t.Errorf("pg_dump must connect through the old installation's data directory %s, got %s", oldDataDir, got)
	}

	restoreArgs := writeArgsRecorder(t, filepaths.PgRestoreBinaryExecutablePath())
	if err := runRestoreUsingList(context.Background(), &RunningDBInstanceInfo{Port: 9193, Database: "steampipe"}, "list"); err != nil {
		t.Fatal(err)
	}
	if got := recordedHost(t, restoreArgs); got != filepaths.GetDataLocation() {
		t.Errorf("pg_restore must connect through the new installation's data directory %s, got %s", filepaths.GetDataLocation(), got)
	}
}

func TestLibpqSocketDir_CommaPathGoesThroughPrivateLink(t *testing.T) {
	dataDir := filepath.Join(shortTempDir(t), "a,b", "data")
	dir, cleanup, err := libpqSocketDir(dataDir, 9193)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(dir, ",") {
		t.Errorf("libpq splits a host on commas: %s", dir)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Errorf("the link directory must be private, got %v", info.Mode().Perm())
	}
	target, err := os.Readlink(rootSocketPath(dir, 9193))
	if err != nil {
		t.Fatal(err)
	}
	if target != rootSocketPath(dataDir, 9193) {
		t.Errorf("the link points at %s, want %s", target, rootSocketPath(dataDir, 9193))
	}
	cleanup()
	if _, err := os.Stat(dir); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("cleanup must remove the link directory, got %v", err)
	}
}

func TestPostgresSocketArgs_DoublesQuotesAndKeepsCommas(t *testing.T) {
	got, err := postgresSocketArgs(`/d/a,b"c`, 9193)
	if err != nil {
		t.Fatal(err)
	}
	if want := `unix_socket_directories="/d/a,b""c"`; got[1] != want {
		t.Errorf("got %s, want %s", got[1], want)
	}
}

func TestLibpqSocketDir_CommaInTempDirUsesSlashTmp(t *testing.T) {
	t.Setenv("TMPDIR", "/x,y")
	dir, cleanup, err := libpqSocketDir("/data/a,b", 9193)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if strings.Contains(dir, ",") || filepath.Dir(dir) != "/tmp" {
		t.Errorf("the link directory must be in /tmp, got %s", dir)
	}
}

func TestLibpqSocketDir_LongTempDirUsesSlashTmp(t *testing.T) {
	longTmp := filepath.Join(shortTempDir(t), strings.Repeat("t", 90))
	if err := os.Mkdir(longTmp, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TMPDIR", longTmp)
	dir, cleanup, err := libpqSocketDir("/data/a,b", 9193)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if err := checkSocketPathLength(dir, 9193); err != nil {
		t.Errorf("the link to the socket must fit the OS limit: %v", err)
	}
}

func TestLibpqSocketDir_PathWithoutCommaIsUsedAsIs(t *testing.T) {
	dir, cleanup, err := libpqSocketDir("/data", 9193)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if dir != "/data" {
		t.Errorf("got %s", dir)
	}
}

func TestSetRootSocketHost_KeepsCommaInPath(t *testing.T) {
	config, err := pgx.ParseConfig("port=9193 user=root sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	setRootSocketHost(config, "/Users/me/a,b/data")
	if config.Host != "/Users/me/a,b/data" || len(config.Fallbacks) != 0 {
		t.Errorf("got host %q with %d fallbacks", config.Host, len(config.Fallbacks))
	}
}

func shortTempDir(t *testing.T) string {
	t.Helper()
	// unix socket paths are limited to about 100 bytes, which t.TempDir can exceed
	dir, err := os.MkdirTemp("/tmp", "sp")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestCheckRootSocket(t *testing.T) {
	dir := shortTempDir(t)

	if err := checkRootSocket(dir, 9193); !errors.Is(err, errStartedByOlderVersion) {
		t.Errorf("a running service with no socket in the data directory was started by an older version, got %v", err)
	}

	l, err := net.Listen("unix", rootSocketPath(dir, 9193))
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if err := checkRootSocket(dir, 9193); err != nil {
		t.Errorf("the socket is there: %v", err)
	}
}

func TestWaitForRootSocket(t *testing.T) {
	dir := shortTempDir(t)
	originalTimeout := viper.Get(pconstants.ArgDatabaseStartTimeout)
	t.Cleanup(func() { viper.Set(pconstants.ArgDatabaseStartTimeout, originalTimeout) })
	viper.Set(pconstants.ArgDatabaseStartTimeout, 0)

	if err := waitForRootSocket(context.Background(), dir, 9193); err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Errorf("want a time-out while the socket is missing, got %v", err)
	}

	l, err := net.Listen("unix", rootSocketPath(dir, 9193))
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if err := waitForRootSocket(context.Background(), dir, 9193); err != nil {
		t.Errorf("the socket is there: %v", err)
	}
}

func TestStartDatabaseInLocation_TooLongSocketPathLeavesOldDataUntouched(t *testing.T) {
	installDir := filepath.Join(t.TempDir(), strings.Repeat("x", 100), ".steampipe")
	originalInstallDir := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = originalInstallDir })
	app_specific.InstallDir = installDir
	oldLocation := filepath.Join(installDir, "db", "14.19.0")
	oldDataDir := filepath.Join(oldLocation, "data")
	if err := os.MkdirAll(oldDataDir, 0700); err != nil {
		t.Fatal(err)
	}
	hbaPath := filepath.Join(oldDataDir, pgHbaFileName)
	if err := os.WriteFile(hbaPath, []byte(olderSteampipeHba), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := startDatabaseInLocation(context.Background(), oldLocation)
	if err == nil {
		t.Fatal("a socket path over the limit must be refused")
	}

	got, readErr := os.ReadFile(hbaPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != olderSteampipeHba {
		t.Errorf("the old data directory's pg_hba.conf must not be rewritten before the socket path is checked, got:\n%s", got)
	}
	if _, statErr := os.Stat(filepath.Join(oldDataDir, pgHbaDirName)); statErr == nil {
		t.Errorf("%s must not be created in the old data directory", pgHbaDirName)
	}
	if !strings.Contains(err.Error(), "Move "+installDir) || !strings.Contains(err.Error(), constants.EnvInstallDir) {
		t.Errorf("the error must say to move the existing install directory and point %s at it: %v", constants.EnvInstallDir, err)
	}
	if strings.Contains(err.Error(), "Set "+constants.EnvInstallDir+" to a shorter path") {
		t.Errorf("an upgrade must not be told to set %s to a new path: %v", constants.EnvInstallDir, err)
	}
}

func TestMoveDataFromSameMajorInstall_TooLongSocketPathGivesNoOldInstallAdvice(t *testing.T) {
	installDir := filepath.Join(t.TempDir(), strings.Repeat("x", 100), ".steampipe")
	originalInstallDir := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = originalInstallDir })
	app_specific.InstallDir = installDir
	oldLocation := filepath.Join(installDir, "db", "14.19.0")
	if err := os.MkdirAll(filepath.Join(oldLocation, "data"), 0700); err != nil {
		t.Fatal(err)
	}

	err := moveDataFromSameMajorInstall(context.Background(), oldLocation, filepaths.GetDataLocation())
	if !errors.Is(err, errSocketPathTooLong) {
		t.Fatalf("want the too-long socket path error, got %v", err)
	}
	// moving the old installation aside would start an empty database and leave the data behind
	if strings.Contains(err.Error(), "new, empty database") || strings.Contains(err.Error(), "run the older version directly") {
		t.Errorf("the error must only say to move the install directory: %v", err)
	}
}

func TestStartDatabaseInLocation_TooLongNewSocketPathLeavesOldDataUntouched(t *testing.T) {
	// the old data directory's socket path fits; the new one's does not
	installDir := filepath.Join(t.TempDir(), strings.Repeat("x", 100), ".steampipe")
	originalInstallDir := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = originalInstallDir })
	app_specific.InstallDir = installDir
	oldLocation := filepath.Join(shortTempDir(t), "db", "14.19.0")
	oldDataDir := filepath.Join(oldLocation, "data")
	if err := os.MkdirAll(oldDataDir, 0700); err != nil {
		t.Fatal(err)
	}
	hbaPath := filepath.Join(oldDataDir, pgHbaFileName)
	if err := os.WriteFile(hbaPath, []byte(olderSteampipeHba), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := startDatabaseInLocation(context.Background(), oldLocation)
	if !errors.Is(err, errSocketPathTooLong) {
		t.Fatalf("a new data directory whose socket path is over the limit must be refused before the old data is touched, got %v", err)
	}
	if !strings.Contains(err.Error(), "Move "+installDir) {
		t.Errorf("the error must say to move the install directory: %v", err)
	}
	if got, readErr := os.ReadFile(hbaPath); readErr != nil || string(got) != olderSteampipeHba {
		t.Errorf("the old data directory's pg_hba.conf must not be rewritten, got %q, %v", got, readErr)
	}
	if _, statErr := os.Stat(filepath.Join(oldDataDir, pgHbaDirName)); statErr == nil {
		t.Errorf("%s must not be created in the old data directory", pgHbaDirName)
	}
}
