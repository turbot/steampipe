package db_local

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
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
	want := []string{"-c", "unix_socket_directories=/data", "-c", "unix_socket_permissions=0700"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPostgresSocketArgs_RefusesTooLongPath(t *testing.T) {
	dataDir := "/" + strings.Repeat("d", maxSocketPathLength())
	got, err := postgresSocketArgs(dataDir, 9193)
	if err == nil {
		t.Fatalf("a socket path longer than %d bytes must be refused, got %v", maxSocketPathLength(), got)
	}
	if !strings.Contains(err.Error(), constants.EnvInstallDir) {
		t.Errorf("the error must say to set %s: %v", constants.EnvInstallDir, err)
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
	// a short install directory: t.TempDir can make the socket path too long
	app_specific.InstallDir = filepath.Join(shortTempDir(t), ".steampipe")
	cmd, err := createCmd(context.Background(), 9193, []string{"127.0.0.1"})
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Join(cmd.Args, " ")
	for _, setting := range []string{"unix_socket_directories=" + filepaths.GetDataLocation(), "unix_socket_permissions=0700"} {
		if !strings.Contains(args, setting) {
			t.Errorf("service start is missing %s: %s", setting, args)
		}
	}
}

func TestRootConnectionParams_UsesSocketWithoutSSL(t *testing.T) {
	got := rootConnectionParams("/data", 9193)
	if got["host"] != "/data" || got["port"] != "9193" || got["user"] != constants.DatabaseSuperUser || got["sslmode"] != "disable" {
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

func TestBackupAndRestoreConnectThroughTheirDataDirectory(t *testing.T) {
	app_specific.InstallDir = filepath.Join(t.TempDir(), ".steampipe")
	oldDataDir := filepath.Join(t.TempDir(), "14.19.0", "data")

	dump := strings.Join(backupCmd(context.Background(), &pgRunningInfo{dataDir: oldDataDir, port: 9876, dbName: "steampipe"}).Args, " ")
	if !strings.Contains(dump, "--host="+oldDataDir+" ") {
		t.Errorf("pg_dump must connect through the old installation's data directory: %s", dump)
	}

	restore := strings.Join(restoreCmd(context.Background(), &RunningDBInstanceInfo{Port: 9193, Database: "steampipe"}, "list").Args, " ")
	if !strings.Contains(restore, "--host="+filepaths.GetDataLocation()+" ") {
		t.Errorf("pg_restore must connect through the new installation's data directory: %s", restore)
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
	viper.Set(pconstants.ArgDatabaseStartTimeout, 0)
	t.Cleanup(func() { viper.Set(pconstants.ArgDatabaseStartTimeout, nil) })

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
