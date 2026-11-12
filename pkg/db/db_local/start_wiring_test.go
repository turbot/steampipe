package db_local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

// every place Steampipe starts a postgres server must write pg_hba.conf first, never start postgres if that
// write fails, and pass the socket settings. Each case runs the real code path with a stand-in postgres, a script
// that records its command line, a copy of the pg_hba.conf in its data folder when it started, and that it ran.

// writeRecordingPostgres writes a stand-in postgres at binaryPath and returns the folder it records into
func writeRecordingPostgres(t *testing.T, binaryPath string) string {
	t.Helper()
	recordDir := t.TempDir()
	script := fmt.Sprintf(`#!/bin/sh
dir=""
prev=""
for arg in "$@"; do
  if [ "$prev" = "-D" ]; then dir="$arg"; fi
  prev="$arg"
done
printf '%%s\n' "$@" > %[1]s/args
if [ -f "$dir/pg_hba.conf" ]; then cp "$dir/pg_hba.conf" %[1]s/hba; fi
touch %[1]s/done
exec sleep 60
`, recordDir)
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binaryPath, []byte(script), 0700); err != nil { //nolint:gosec // G306: the stand-in must be executable
		t.Fatal(err)
	}
	return recordDir
}

// assertRecordedStart waits for the stand-in postgres to have started, then checks what it recorded
func assertRecordedStart(t *testing.T, recordDir string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(filepath.Join(recordDir, "done")); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the stand-in postgres never started")
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Run("writes pg_hba.conf before postgres starts", func(t *testing.T) {
		hba, err := os.ReadFile(filepath.Join(recordDir, "hba"))
		if err != nil {
			t.Fatalf("pg_hba.conf did not exist when postgres started: %v", err)
		}
		if !strings.HasPrefix(string(hba), strings.SplitN(constants.PgHbaHeader, "\n", 2)[0]) {
			t.Errorf("pg_hba.conf was not written by Steampipe:\n%s", hba)
		}
	})
	t.Run("passes the socket settings", func(t *testing.T) {
		args, err := os.ReadFile(filepath.Join(recordDir, "args"))
		if err != nil {
			t.Fatal(err)
		}
		for _, setting := range []string{"unix_socket_directories=/tmp", "unix_socket_permissions=0700"} {
			if !strings.Contains(string(args), setting) {
				t.Errorf("postgres was started without %s:\n%s", setting, args)
			}
		}
	})
}

func TestStartPostgresProcess_WritesPgHbaAndSetsSocket(t *testing.T) {
	setCleanupInstallDir(t)
	recordDir := writeRecordingPostgres(t, filepaths.GetPostgresBinaryExecutablePath())
	_ = filepaths.GetDataLocation() // creates the data folder

	cmd, err := startPostgresProcess(context.Background(), []string{"127.0.0.1"}, 9193, constants.InvokerService)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})

	assertRecordedStart(t, recordDir)
}

func TestStartServiceForInstall_WritesPgHbaAndSetsSocket(t *testing.T) {
	setCleanupInstallDir(t)
	recordDir := writeRecordingPostgres(t, filepaths.GetPostgresBinaryExecutablePath())
	_ = filepaths.GetDataLocation() // creates the data folder

	process, err := startServiceForInstall(9193)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { killAndReap(int(process.Pid)) })

	assertRecordedStart(t, recordDir)
}

func TestStartDatabaseInLocation_WritesPgHbaAndSetsSocket(t *testing.T) {
	setCleanupInstallDir(t)
	location := filepath.Join(app_specific.InstallDir, "db", "13.0.0")
	recordDir := writeRecordingPostgres(t, filepath.Join(location, "postgres", "bin", "postgres"))
	if err := os.MkdirAll(filepath.Join(location, "data"), 0700); err != nil {
		t.Fatal(err)
	}

	// the stand-in never opens a socket, so starting waits for one until the context ends; that kills the process
	previousTimeout := viper.Get(pconstants.ArgDatabaseStartTimeout)
	t.Cleanup(func() { viper.Set(pconstants.ArgDatabaseStartTimeout, previousTimeout) })
	viper.Set(pconstants.ArgDatabaseStartTimeout, 60)
	ctx, cancel := context.WithCancel(context.Background())
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		_, _ = startDatabaseInLocation(ctx, location)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-finished:
		case <-time.After(15 * time.Second):
			t.Error("startDatabaseInLocation did not return after its context was cancelled")
		}
	})

	assertRecordedStart(t, recordDir)
}

// A pg_hba.conf that cannot be written must stop postgres from being started at all. Writing it after the start
// would let a server run, briefly, with whatever rules the data folder held. A file stands where the pg_hba.conf.d
// folder belongs, so the write fails; the stand-in postgres records if it was ever launched.
func assertNotStartedWhenPgHbaCannotBeWritten(t *testing.T, recordDir, dataLocation string, start func() error) {
	t.Helper()
	if err := os.MkdirAll(dataLocation, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataLocation, pgHbaDirName), nil, 0600); err != nil {
		t.Fatal(err)
	}

	if err := start(); err == nil {
		t.Fatal("expected an error when pg_hba.conf cannot be written")
	}

	// a stand-in that was launched records itself within milliseconds
	time.Sleep(time.Second)
	if _, err := os.Stat(filepath.Join(recordDir, "done")); err == nil {
		t.Error("postgres was started although pg_hba.conf could not be written first")
	}
}

func TestStartPostgresProcess_DoesNotStartWhenPgHbaCannotBeWritten(t *testing.T) {
	setCleanupInstallDir(t)
	recordDir := writeRecordingPostgres(t, filepaths.GetPostgresBinaryExecutablePath())

	assertNotStartedWhenPgHbaCannotBeWritten(t, recordDir, filepaths.GetDataLocation(), func() error {
		cmd, err := startPostgresProcess(context.Background(), []string{"127.0.0.1"}, 9193, constants.InvokerService)
		if cmd != nil {
			t.Cleanup(func() {
				_ = cmd.Process.Kill()
				_ = cmd.Wait()
			})
		}
		return err
	})
}

func TestStartServiceForInstall_DoesNotStartWhenPgHbaCannotBeWritten(t *testing.T) {
	setCleanupInstallDir(t)
	recordDir := writeRecordingPostgres(t, filepaths.GetPostgresBinaryExecutablePath())

	assertNotStartedWhenPgHbaCannotBeWritten(t, recordDir, filepaths.GetDataLocation(), func() error {
		process, err := startServiceForInstall(9193)
		if process != nil {
			t.Cleanup(func() { killAndReap(int(process.Pid)) })
		}
		return err
	})
}

func TestStartDatabaseInLocation_DoesNotStartWhenPgHbaCannotBeWritten(t *testing.T) {
	setCleanupInstallDir(t)
	location := filepath.Join(app_specific.InstallDir, "db", "13.0.0")
	recordDir := writeRecordingPostgres(t, filepath.Join(location, "postgres", "bin", "postgres"))

	assertNotStartedWhenPgHbaCannotBeWritten(t, recordDir, filepath.Join(location, "data"), func() error {
		// cancelled at the end so that a process started by mistake is killed with it
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		_, err := startDatabaseInLocation(ctx, location)
		return err
	})
}

// killAndReap stops a child process that was started inside production code, so that it does not stay behind as a
// zombie until the test binary exits
func killAndReap(pid int) {
	if p, err := os.FindProcess(pid); err == nil {
		_ = p.Kill()
		_, _ = p.Wait()
	}
}
