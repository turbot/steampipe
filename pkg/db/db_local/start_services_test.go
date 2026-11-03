package db_local

import (
	"context"
	"os"
	"os/exec"
	"slices"
	"testing"
	"time"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/turbot/pipe-fittings/v2/app_specific"
)

const fakePostgresArg = "application_name=steampipe"

// Run as a child process by TestFindSteampipePostgresInstancesSkipsExitedProcess,
// which gives it a command line that looks like a steampipe postgres service.
func TestFakeSteampipePostgresProcess(t *testing.T) {
	if !slices.Contains(os.Args, fakePostgresArg) {
		t.Skip("only runs as a child process")
	}
	time.Sleep(time.Minute)
}

// A process that exits between listing and reading its command line must not
// stop the scan: `service stop --force` relies on it to find running instances.
func TestFindSteampipePostgresInstancesSkipsExitedProcess(t *testing.T) {
	previousAppName := app_specific.AppName
	app_specific.AppName = "steampipe"
	t.Cleanup(func() { app_specific.AppName = previousAppName })

	exitedCmd := exec.Command("true")
	if err := exitedCmd.Run(); err != nil {
		t.Fatalf("failed to run child process: %v", err)
	}

	postgresCmd := exec.Command(os.Args[0], "-test.run=^TestFakeSteampipePostgresProcess$", fakePostgresArg) //nolint:gosec // G204: os.Args[0] is this test binary, re-run with fixed arguments
	postgresCmd.Args[0] = "postgres"
	if err := postgresCmd.Start(); err != nil {
		t.Fatalf("failed to start fake postgres process: %v", err)
	}
	t.Cleanup(func() {
		_ = postgresCmd.Process.Kill()
		_ = postgresCmd.Wait()
	})

	exited := &psutils.Process{Pid: int32(exitedCmd.Process.Pid)}     //nolint:gosec // G115: test child PID
	postgres := &psutils.Process{Pid: int32(postgresCmd.Process.Pid)} //nolint:gosec // G115: test child PID
	listProcesses := func(context.Context) ([]*psutils.Process, error) {
		return []*psutils.Process{exited, postgres}, nil
	}

	instances, err := findSteampipePostgresInstances(context.Background(), listProcesses)
	if err != nil {
		t.Fatalf("expected the exited process to be skipped, got error: %v", err)
	}
	if len(instances) != 1 || instances[0].Pid != postgres.Pid {
		t.Errorf("expected only the fake postgres process (pid %d), got %v", postgres.Pid, instances)
	}
}
