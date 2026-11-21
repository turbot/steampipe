package db_local

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/spf13/viper"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

const fakePostgresArg = "application_name=steampipe"

// Run as a child process by the TestFindSteampipePostgresInstances* tests,
// which give it a command line that looks like a steampipe postgres service.
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
	// On Linux the command line of a just-started child can read as empty
	// for a short time, so wait until it shows the arguments.
	deadline := time.Now().Add(10 * time.Second)
	for {
		cmdline, err := postgres.CmdlineSlice()
		if err == nil && slices.Contains(cmdline, fakePostgresArg) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("fake postgres process never showed its command line, last read %q (err %v)", cmdline, err)
		}
		time.Sleep(time.Millisecond)
	}

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

// A process that has only just started can briefly report an empty command
// line on Linux. The scan must still find it, and must not match a process
// whose command line stays empty.
func TestFindSteampipePostgresInstancesFindsProcessWithBrieflyEmptyCmdline(t *testing.T) {
	previousAppName := app_specific.AppName
	app_specific.AppName = "steampipe"
	t.Cleanup(func() { app_specific.AppName = previousAppName })

	postgresCmd := exec.Command(os.Args[0], "-test.run=^TestFakeSteampipePostgresProcess$", fakePostgresArg) //nolint:gosec // G204: os.Args[0] is this test binary, re-run with fixed arguments
	postgresCmd.Args[0] = "postgres"
	if err := postgresCmd.Start(); err != nil {
		t.Fatalf("failed to start fake postgres process: %v", err)
	}
	t.Cleanup(func() {
		_ = postgresCmd.Process.Kill()
		_ = postgresCmd.Wait()
	})

	postgres := &psutils.Process{Pid: int32(postgresCmd.Process.Pid)} //nolint:gosec // G115: test child PID
	alwaysEmpty := &psutils.Process{Pid: int32(os.Getpid())}          //nolint:gosec // G115: this test's own PID

	previousReader := readProcessCmdline
	t.Cleanup(func() { readProcessCmdline = previousReader })
	postgresReads := 0
	readProcessCmdline = func(ctx context.Context, p *psutils.Process) ([]string, error) {
		if p.Pid == alwaysEmpty.Pid {
			return []string{}, nil
		}
		postgresReads++
		if postgresReads <= 2 {
			return []string{}, nil
		}
		return previousReader(ctx, p)
	}

	listProcesses := func(context.Context) ([]*psutils.Process, error) {
		return []*psutils.Process{alwaysEmpty, postgres}, nil
	}

	instances, err := findSteampipePostgresInstances(context.Background(), listProcesses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(instances) != 1 || instances[0].Pid != postgres.Pid {
		t.Errorf("expected only the fake postgres process (pid %d), got %v", postgres.Pid, instances)
	}
}

// A process with an empty command line that started before the recently-started window may be a kernel thread
// or a zombie, which always read empty. It must be read once and not retried.
func TestFindSteampipePostgresInstancesDoesNotRetryOldProcessWithEmptyCmdline(t *testing.T) {
	// the first process of the system has normally run for far longer than the window; its age is checked, not assumed
	oldProcess := &psutils.Process{Pid: 1}
	createdMs, err := oldProcess.CreateTime()
	if err != nil || time.Since(time.UnixMilli(createdMs)) < 20*time.Second {
		t.Skipf("process 1 has not run for 20 seconds (created %d, err %v)", createdMs, err)
	}

	previousReader := readProcessCmdline
	t.Cleanup(func() { readProcessCmdline = previousReader })
	reads := 0
	readProcessCmdline = func(ctx context.Context, p *psutils.Process) ([]string, error) {
		reads++
		return []string{}, nil
	}

	listProcesses := func(context.Context) ([]*psutils.Process, error) {
		return []*psutils.Process{oldProcess}, nil
	}

	instances, err := findSteampipePostgresInstances(context.Background(), listProcesses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(instances) != 0 {
		t.Errorf("expected no instances, got %v", instances)
	}
	if reads != 1 {
		t.Errorf("a process older than %s should be read once, was read %d times", recentlyStartedWindow, reads)
	}
}

// A cancelled context stops the wait for a recently started process to show its command line, and the matches
// found so far come back with the context's error.
func TestFindSteampipePostgresInstancesReturnsContextErrorWhenCancelled(t *testing.T) {
	startChild := func() *exec.Cmd {
		cmd := exec.Command("sleep", "60")
		if err := cmd.Start(); err != nil {
			t.Fatalf("failed to start child process: %v", err)
		}
		t.Cleanup(func() {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		})
		return cmd
	}
	matching := startChild()
	justStarted := startChild()
	matchingProcess := &psutils.Process{Pid: int32(matching.Process.Pid)}       //nolint:gosec // G115: test child PID
	justStartedProcess := &psutils.Process{Pid: int32(justStarted.Process.Pid)} //nolint:gosec // G115: test child PID

	// the process remembers its start time once read, so the scan can still read it after the context ends
	if _, err := justStartedProcess.CreateTime(); err != nil {
		t.Fatalf("failed to read the start time of the child process: %v", err)
	}

	previousAppName := app_specific.AppName
	app_specific.AppName = "steampipe"
	t.Cleanup(func() { app_specific.AppName = previousAppName })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	previousReader := readProcessCmdline
	t.Cleanup(func() { readProcessCmdline = previousReader })
	readProcessCmdline = func(_ context.Context, p *psutils.Process) ([]string, error) {
		if p.Pid == matchingProcess.Pid {
			return []string{"postgres", "-c", fakePostgresArg}, nil
		}
		// the process still reads empty, and the context ends before it is read again
		cancel()
		return []string{}, nil
	}

	listProcesses := func(context.Context) ([]*psutils.Process, error) {
		return []*psutils.Process{matchingProcess, justStartedProcess}, nil
	}

	instances, err := findSteampipePostgresInstances(ctx, listProcesses)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected the context's error, got %v", err)
	}
	if len(instances) != 1 || instances[0].Pid != matchingProcess.Pid {
		t.Errorf("expected the matches found before the cancel (pid %d), got %v", matchingProcess.Pid, instances)
	}
}

// Postgres runs ssl_passphrase_command through /bin/sh and puts its own command
// line in front of every local user (ps, /proc). So the passphrase must not be
// in the command line, and the command must print it back exactly, without the
// shell acting on any of its characters.
func TestCreateCmdKeepsSSLPassphraseOutOfCommand(t *testing.T) {
	originalInstallDir := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = originalInstallDir })
	// a short install directory: t.TempDir can make the socket path too long
	workDir := shortTempDir(t)
	app_specific.InstallDir = filepath.Join(workDir, "steampipe")

	passphrase := `it's a "pass" $(touch pwned) ` + "`touch pwned` ; touch pwned \\n %p %%"
	viper.Set(pconstants.ArgDatabaseSSLPassword, passphrase)
	t.Cleanup(func() { viper.Set(pconstants.ArgDatabaseSSLPassword, "") })

	// make createCmd also take the branch that sets OPENSSL_CONF
	t.Setenv("OPENSSL_CONF", "")
	os.Unsetenv("OPENSSL_CONF")

	cmd, err := createCmd(context.Background(), 9193, []string{"localhost"})
	if err != nil {
		t.Fatal(err)
	}

	var passphraseCommand string
	for _, arg := range cmd.Args {
		if strings.Contains(arg, "pwned") {
			t.Errorf("expected the passphrase to be kept out of the postgres command line, got argument %q", arg)
		}
		if value, ok := strings.CutPrefix(arg, "ssl_passphrase_command="); ok {
			passphraseCommand = value
		}
	}
	if passphraseCommand == "" {
		t.Fatalf("expected an ssl_passphrase_command argument, got %v", cmd.Args)
	}
	if !slices.Contains(cmd.Env, constants.EnvInstallDir+"="+app_specific.InstallDir) {
		t.Errorf("expected the postgres environment to keep %s", constants.EnvInstallDir)
	}

	// run the command the way postgres does: replace %% with %, then run it with /bin/sh
	shell := exec.Command("/bin/sh", "-c", strings.ReplaceAll(passphraseCommand, "%%", "%")) //nolint:gosec // G204: runs the passphrase command that createCmd built, as postgres would
	shell.Env = cmd.Env
	shell.Dir = workDir
	output, err := shell.Output()
	if err != nil {
		t.Fatalf("passphrase command %q failed: %v", passphraseCommand, err)
	}
	if got := strings.TrimSuffix(string(output), "\n"); got != passphrase {
		t.Errorf("expected the passphrase command to print %q, got %q", passphrase, got)
	}
	if _, err := os.Stat(filepath.Join(workDir, "pwned")); err == nil {
		t.Errorf("the shell ran a command from the passphrase")
	}
}
