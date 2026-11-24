package db_local

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

// setStateTestDirs points service state at a scratch install and sets the
// application name the database command line is matched against.
func setStateTestDirs(t *testing.T) {
	t.Helper()
	prevDir := app_specific.InstallDir
	prevName := app_specific.AppName
	app_specific.InstallDir = t.TempDir()
	app_specific.AppName = "steampipe"
	t.Cleanup(func() {
		app_specific.InstallDir = prevDir
		app_specific.AppName = prevName
	})
}

func writeStateForPid(t *testing.T, pid int) {
	t.Helper()
	info := newRunningDBInstanceInfo(&exec.Cmd{Process: &os.Process{Pid: pid}}, []string{"localhost"}, 9193, "steampipe", "password", constants.InvokerService)
	if err := info.Save(); err != nil {
		t.Fatalf("save running info: %v", err)
	}
}

func writePostmasterPid(t *testing.T, pid int) {
	t.Helper()
	// the first line is the only one GetState reads
	if err := os.WriteFile(filepaths.GetPostmasterPidLocation(), []byte(fmt.Sprintf("%d\n", pid)), 0600); err != nil {
		t.Fatalf("write postmaster.pid: %v", err)
	}
}

// procStatState is the process state from /proc/pid/stat. The command name
// can contain spaces, so the state is the field after the last ')'.
func procStatState(stat []byte) string {
	s := string(stat)
	i := strings.LastIndex(s, ")")
	if i < 0 || i+2 >= len(s) {
		return ""
	}
	fields := strings.Fields(s[i+1:])
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// zombiePid starts a child and kills it without reaping, so it stays a zombie
// of this test until cleanup.
func zombiePid(t *testing.T) int {
	t.Helper()
	cmd := exec.Command("sleep", "60")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start child: %v", err)
	}
	pid := cmd.Process.Pid
	if err := cmd.Process.Kill(); err != nil {
		t.Fatalf("kill child: %v", err)
	}
	t.Cleanup(func() { _ = cmd.Wait() })

	deadline := time.Now().Add(2 * time.Second)
	for {
		stat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if err == nil && procStatState(stat) == "Z" {
			return pid
		}
		if time.Now().After(deadline) {
			t.Fatalf("pid %d did not become a zombie (stat %q, err %v)", pid, stat, err)
		}
		time.Sleep(time.Millisecond)
	}
}

func startFakeDatabase(t *testing.T) int {
	t.Helper()
	postgresCmd := exec.Command(os.Args[0], "-test.run=^TestFakeSteampipePostgresProcess$", fakePostgresArg) //nolint:gosec // G204: os.Args[0] is this test binary, re-run with fixed arguments
	postgresCmd.Args[0] = "postgres"
	if err := postgresCmd.Start(); err != nil {
		t.Fatalf("start fake database: %v", err)
	}
	t.Cleanup(func() {
		_ = postgresCmd.Process.Kill()
		_ = postgresCmd.Wait()
	})

	postgres := &psutils.Process{Pid: int32(postgresCmd.Process.Pid)} //nolint:gosec // G115: test child PID
	deadline := time.Now().Add(10 * time.Second)
	for {
		cmdline, err := postgres.CmdlineSlice()
		if err == nil && slices.Contains(cmdline, fakePostgresArg) {
			return postgresCmd.Process.Pid
		}
		if time.Now().After(deadline) {
			t.Fatalf("fake database never showed its command line, last read %q (err %v)", cmdline, err)
		}
		time.Sleep(time.Millisecond)
	}
}

// A database that has exited but is still a zombie must not count as running,
// including on the next status check, when only postmaster.pid is left.
func TestGetState_ZombieDatabasePidIsNotRunning(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("the zombie state is read from /proc")
	}
	setStateTestDirs(t)
	pid := zombiePid(t)
	if !utils.PidExists(pid) {
		t.Fatalf("pipe-fittings PidExists(%d) = false for a zombie; the reported check does not treat it as running here", pid)
	}
	writeStateForPid(t, pid)
	writePostmasterPid(t, pid)

	state, err := GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	if state != nil {
		t.Fatalf("GetState reported zombie pid %d as a running database", pid)
	}

	state, err = GetState()
	if err != nil {
		t.Fatalf("GetState after the state file was removed: %v", err)
	}
	if state != nil {
		t.Fatalf("GetState reported zombie pid %d as a running database on the second check", pid)
	}
}

// A pid recorded for the database can be reused by an unrelated process after
// a restart. That process is not the database.
func TestGetState_ReusedPidIsNotRunning(t *testing.T) {
	setStateTestDirs(t)
	cmd := exec.Command("sleep", "60")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start unrelated process: %v", err)
	}
	t.Cleanup(func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	})
	pid := cmd.Process.Pid
	if !utils.PidExists(pid) {
		t.Fatalf("pipe-fittings PidExists(%d) = false for a live process", pid)
	}
	writeStateForPid(t, pid)

	state, err := GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	if state != nil {
		t.Fatalf("GetState reported unrelated pid %d as a running database", pid)
	}
}

// A live steampipe database with no service state file is still the unknown
// state the startup path refuses to touch. Only a dead or unrelated pid is stale.
func TestGetState_LiveDatabaseWithoutStateFileIsUnknown(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake database is a re-exec of this test binary")
	}
	setStateTestDirs(t)
	pid := startFakeDatabase(t)
	writePostmasterPid(t, pid)

	state, err := GetState()
	if err == nil {
		t.Fatalf("GetState = %+v, want an unknown-state error", state)
	}
	if !strings.Contains(err.Error(), "unknown state") {
		t.Fatalf("GetState error = %v, want unknown state", err)
	}
}

// A live process whose command line is the steampipe database still counts as running.
func TestGetState_RunningSteampipeDatabaseIsRunning(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake database is a re-exec of this test binary")
	}
	setStateTestDirs(t)
	pid := startFakeDatabase(t)
	writeStateForPid(t, pid)

	state, err := GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	if state == nil || state.Pid != pid {
		t.Fatalf("GetState = %+v, want the database pid %d", state, pid)
	}
}

// A readable postgres command line without application_name=steampipe belongs
// to another program, so the stale state files are removed.
func TestGetState_OtherPostgresIsNotRunning(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake database is a re-exec of this test binary")
	}
	setStateTestDirs(t)
	pid := startFakeDatabase(t)
	writeStateForPid(t, pid)
	writePostmasterPid(t, pid)

	prev := readProcessCmdline
	readProcessCmdline = func(context.Context, *psutils.Process) ([]string, error) {
		return []string{"/usr/lib/postgresql/16/bin/postgres", "-D", "data"}, nil
	}
	t.Cleanup(func() { readProcessCmdline = prev })

	state, err := GetState()
	if err != nil {
		t.Fatalf("GetState: %v", err)
	}
	if state != nil {
		t.Fatalf("GetState = %+v, want nil for a postgres that is not steampipe", state)
	}
	if _, err := os.Stat(filepaths.RunningInfoFilePath()); err == nil {
		t.Fatal("state file kept for a postgres that is not steampipe")
	}
}

// State is kept when the live process cannot be read, and when a space in the
// install path splits the postgres binary out of argv[0]. Both still report
// the service as running: the first because it is not known to be something
// else, the second because it is the steampipe database.
func TestGetState_UnconfirmedLiveDatabaseKeepsState(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake database is a re-exec of this test binary")
	}
	cases := map[string]func(*psutils.Process) ([]string, error){
		"command line read fails": func(*psutils.Process) ([]string, error) {
			return nil, fmt.Errorf("permission denied")
		},
		"command line split at a space in the install path": func(*psutils.Process) ([]string, error) {
			return []string{"/Users/me/My", "Stuff/bin/postgres", "-D", "data", "application_name=steampipe"}, nil
		},
	}
	for name, read := range cases {
		t.Run(name, func(t *testing.T) {
			setStateTestDirs(t)
			pid := startFakeDatabase(t)
			writeStateForPid(t, pid)
			writePostmasterPid(t, pid)

			prev := readProcessCmdline
			readProcessCmdline = func(_ context.Context, p *psutils.Process) ([]string, error) { return read(p) }
			t.Cleanup(func() { readProcessCmdline = prev })

			state, err := GetState()
			if err != nil {
				t.Fatalf("GetState: %v", err)
			}
			if state == nil || state.Pid != pid {
				t.Fatalf("GetState = %+v, want the live pid %d reported as running", state, pid)
			}
			if _, err := os.Stat(filepaths.RunningInfoFilePath()); err != nil {
				t.Fatalf("state file removed for a live database: %v", err)
			}
		})
	}
}
