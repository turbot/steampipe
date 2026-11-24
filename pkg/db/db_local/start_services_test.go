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

// gopsutil on macOS splits `ps -o command` on spaces, and on Windows it splits
// the command line on spaces too. A space in the install path leaves the
// postgres binary in a later argument. That process is still the steampipe
// database. Another postgres, without application_name=steampipe, is not.
func TestFindSteampipePostgresInstancesMatchesInstallPathSplitAtSpace(t *testing.T) {
	previousAppName := app_specific.AppName
	app_specific.AppName = "steampipe"
	t.Cleanup(func() { app_specific.AppName = previousAppName })

	split := &psutils.Process{Pid: 4242}
	otherPostgres := &psutils.Process{Pid: 4243}
	windowsSplit := &psutils.Process{Pid: 4244}

	previousReader := readProcessCmdline
	t.Cleanup(func() { readProcessCmdline = previousReader })
	readProcessCmdline = func(_ context.Context, p *psutils.Process) ([]string, error) {
		switch p.Pid {
		case split.Pid:
			return []string{"/Users/me/My", "Stuff/bin/postgres", "-D", "data", "application_name=steampipe"}, nil
		case windowsSplit.Pid:
			return []string{`C:\Program`, `Files\Steampipe\db\postgres\bin\postgres.exe`, "-D", "data", "application_name=steampipe"}, nil
		default:
			return []string{"/usr/lib/postgresql/16/bin/postgres", "-D", "data"}, nil
		}
	}

	instances, err := findSteampipePostgresInstances(context.Background(), func(context.Context) ([]*psutils.Process, error) {
		return []*psutils.Process{split, otherPostgres, windowsSplit}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := map[int32]bool{}
	for _, instance := range instances {
		got[instance.Pid] = true
	}
	if !got[split.Pid] || !got[windowsSplit.Pid] || got[otherPostgres.Pid] || len(instances) != 2 {
		t.Fatalf("instances = %v, want the split-path databases %d and %d only", instances, split.Pid, windowsSplit.Pid)
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

// The service password is written into ALTER USER as a string literal, so a quote or a
// backslash in it must stay part of the password, whether or not standard_conforming_strings is on.
func TestQuoteLiteral(t *testing.T) {
	tests := map[string]string{
		"plain":                          `'plain'`,
		"it's":                           `'it''s'`,
		`back\slash`:                     `E'back\\slash'`,
		`x'; copy (select 1) to 'f'; --`: `'x''; copy (select 1) to ''f''; --'`,
		`it's a \'quoted\' pass`:         `E'it''s a \\''quoted\\'' pass'`,
	}
	for in, want := range tests {
		if got := quoteLiteral(in); got != want {
			t.Errorf("quoteLiteral(%q) = %s, want %s", in, got, want)
		}
	}
}

// startFakeSteampipePostgres starts a child process whose command line looks like a steampipe postgres service
// running on dataDir (no data directory when dataDir is empty). The returned channel closes when it exits.
func startFakeSteampipePostgres(t *testing.T, dataDir string) (*psutils.Process, <-chan struct{}) {
	t.Helper()
	return startFakeSteampipePostgresAt(t, "postgres", "", dataDir)
}

// startFakeSteampipePostgresAt is startFakeSteampipePostgres with the program name (argv[0]) and working directory
// of the process chosen. Real postgres changes into its data directory, so a test of a relative data directory
// starts the process there.
func startFakeSteampipePostgresAt(t *testing.T, argv0, workDir, dataDir string) (*psutils.Process, <-chan struct{}) {
	t.Helper()
	args := []string{"-test.run=^TestFakeSteampipePostgresProcess$", fakePostgresArg}
	if dataDir != "" {
		args = append(args, "-D", dataDir)
	}
	cmd := exec.Command(os.Args[0], args...) //nolint:gosec // G204: os.Args[0] is this test binary, re-run with fixed arguments
	cmd.Args[0] = argv0
	cmd.Dir = workDir
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start fake postgres process: %v", err)
	}
	exited := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(exited)
	}()
	t.Cleanup(func() { _ = cmd.Process.Kill() })

	p := &psutils.Process{Pid: int32(cmd.Process.Pid)} //nolint:gosec // G115: test child PID
	deadline := time.Now().Add(10 * time.Second)
	for {
		cmdline, err := p.CmdlineSlice()
		if err == nil && slices.Contains(cmdline, fakePostgresArg) && (dataDir == "" || slices.Contains(cmdline, dataDir)) {
			return p, exited
		}
		if time.Now().After(deadline) {
			t.Fatalf("fake postgres process never showed its command line, last read %q (err %v)", cmdline, err)
		}
		time.Sleep(time.Millisecond)
	}
}

func useFakeProcessList(t *testing.T, processes ...*psutils.Process) {
	t.Helper()
	previousAppName := app_specific.AppName
	app_specific.AppName = "steampipe"
	t.Cleanup(func() { app_specific.AppName = previousAppName })
	previousList := listAllProcesses
	listAllProcesses = func(context.Context) ([]*psutils.Process, error) { return processes, nil }
	t.Cleanup(func() { listAllProcesses = previousList })
}

// isRunning reports whether the process is still running after a short wait: a process that was just killed takes a
// moment to be reaped.
func isRunning(exited <-chan struct{}) bool {
	select {
	case <-exited:
		return false
	case <-time.After(300 * time.Millisecond):
		return true
	}
}

// Initialising a new install directory stops the leftovers of that directory only: a service running from another
// install directory keeps running (#5158).
func TestKillInstanceInDataDirIfAnyLeavesOtherInstallsRunning(t *testing.T) {
	root := t.TempDir()
	ownDataDir := filepath.Join(root, "a", "db", "14", "data")
	otherDataDir := filepath.Join(root, "b", "db", "14", "data")

	own, ownExited := startFakeSteampipePostgres(t, ownDataDir)
	other, otherExited := startFakeSteampipePostgres(t, otherDataDir)
	noDataDir, noDataDirExited := startFakeSteampipePostgres(t, "")
	useFakeProcessList(t, own, other, noDataDir)

	if !killInstanceInDataDirIfAny(context.Background(), ownDataDir) {
		t.Fatal("expected the leftover of this install directory to be reported as stopped")
	}
	select {
	case <-ownExited:
	case <-time.After(10 * time.Second):
		t.Fatal("the process of this install directory was not stopped")
	}
	if !isRunning(otherExited) {
		t.Error("a service from another install directory was stopped")
	}
	if !isRunning(noDataDirExited) {
		t.Error("a process with no data directory on its command line was stopped")
	}
}

// With nothing of this install directory running, nothing is stopped and nothing is reported.
func TestKillInstanceInDataDirIfAnyStopsNothingWhenNoneAreOurs(t *testing.T) {
	root := t.TempDir()
	other, otherExited := startFakeSteampipePostgres(t, filepath.Join(root, "b", "data"))
	useFakeProcessList(t, other)

	if killInstanceInDataDirIfAny(context.Background(), filepath.Join(root, "a", "data")) {
		t.Error("reported a stop although no process belongs to this install directory")
	}
	if !isRunning(otherExited) {
		t.Error("a service from another install directory was stopped")
	}
}

// `service stop --force` stays machine-wide: it stops every steampipe service, whichever install directory it is from.
func TestKillInstanceIfAnyStopsEveryInstall(t *testing.T) {
	root := t.TempDir()
	first, firstExited := startFakeSteampipePostgres(t, filepath.Join(root, "a", "data"))
	second, secondExited := startFakeSteampipePostgres(t, filepath.Join(root, "b", "data"))
	useFakeProcessList(t, first, second)

	if !killInstanceIfAny(context.Background()) {
		t.Fatal("expected a stop to be reported")
	}
	for name, exited := range map[string]<-chan struct{}{"first": firstExited, "second": secondExited} {
		select {
		case <-exited:
		case <-time.After(10 * time.Second):
			t.Errorf("the %s install's service was not stopped", name)
		}
	}
}

func TestProcessUsesDataDir(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "a", "data")
	link := filepath.Join(root, "link")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "a"), link); err != nil {
		t.Skipf("cannot make a link: %v", err)
	}
	self := &psutils.Process{Pid: int32(os.Getpid())} //nolint:gosec // G115: this test's own PID
	tests := []struct {
		name    string
		cmdline []string
		want    bool
	}{
		{"-D with a space", []string{"postgres", "-D", dir}, true},
		{"-D joined", []string{"postgres", "-D" + dir}, true},
		{"--pgdata", []string{"postgres", "--pgdata", dir}, true},
		{"--pgdata=", []string{"postgres", "--pgdata=" + dir}, true},
		{"unclean path", []string{"postgres", "-D", dir + "/../data/"}, true},
		{"through a link", []string{"postgres", "-D", filepath.Join(link, "data")}, true},
		{"another directory", []string{"postgres", "-D", filepath.Join(root, "b", "data")}, false},
		{"a parent directory", []string{"postgres", "-D", filepath.Dir(dir)}, false},
		{"a child directory", []string{"postgres", "-D", filepath.Join(dir, "base")}, false},
		{"no data directory", []string{"postgres", "-c", "application_name=steampipe"}, false},
		{"-D with no value", []string{"postgres", "-D"}, false},
		{"empty value", []string{"postgres", "-D", ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := processUsesDataDir(context.Background(), self, tt.cmdline, dir); got != tt.want {
				t.Errorf("processUsesDataDir(%q) = %v, want %v", tt.cmdline, got, tt.want)
			}
		})
	}
	if processUsesDataDir(context.Background(), self, []string{"postgres", "-D", dir}, "") {
		t.Error("an empty data directory must match nothing")
	}
}

// A relative install directory (STEAMPIPE_INSTALL_DIR=sp) means a different directory from each working directory:
// two jobs using `sp` in different folders are separate installs, and one must not stop the other's service.
func TestKillInstanceInDataDirIfAnyRelativeInstallDirs(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	jobA := filepath.Join(root, "jobA")
	jobB := filepath.Join(root, "jobB")
	for _, job := range []string{jobA, jobB} {
		if err := os.MkdirAll(filepath.Join(job, "sp", "data"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	serviceA, exitedA := startFakeSteampipePostgresAt(t, "postgres", jobA, filepath.Join("sp", "data"))
	useFakeProcessList(t, serviceA)
	// postgres has changed into its data directory by the time anything scans for it, and not every platform can
	// report a working directory, so the test says what it is
	previousCwdReader := readProcessCwd
	t.Cleanup(func() { readProcessCwd = previousCwdReader })
	readProcessCwd = func(context.Context, *psutils.Process) (string, error) {
		return filepath.Join(jobA, "sp", "data"), nil
	}

	t.Chdir(jobB)
	if killInstanceInDataDirIfAny(context.Background(), filepath.Join("sp", "data")) {
		t.Error("reported a stop for the other job's service")
	}
	if !isRunning(exitedA) {
		t.Fatal("initialising jobB's install stopped jobA's service, which has the same relative install directory")
	}

	t.Chdir(jobA)
	if !killInstanceInDataDirIfAny(context.Background(), filepath.Join("sp", "data")) {
		t.Fatal("expected the leftover of jobA's own install to be stopped")
	}
	select {
	case <-exitedA:
	case <-time.After(10 * time.Second):
		t.Fatal("jobA's own leftover was not stopped")
	}
}

// The upgrade path also clears an install's leftover postgres. An install whose directory name only starts with this
// one's (/sp-10 and /sp-1) is another install, and keeps running.
func TestKillRunningDbInstanceLeavesInstallWithSamePrefixRunning(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ours := filepath.Join(root, "sp-1")
	theirs := filepath.Join(root, "sp-10")
	binary := func(install string) string {
		return filepath.Join(install, "db", "14", "postgres", "bin", "postgres")
	}
	own, ownExited := startFakeSteampipePostgresAt(t, binary(ours), "", filepath.Join(ours, "db", "14", "data"))
	other, otherExited := startFakeSteampipePostgresAt(t, binary(theirs), "", filepath.Join(theirs, "db", "14", "data"))
	useFakeProcessList(t, own, other)
	previousInstallDir := app_specific.InstallDir
	app_specific.InstallDir = ours
	t.Cleanup(func() { app_specific.InstallDir = previousInstallDir })

	if err := killRunningDbInstance(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	select {
	case <-ownExited:
	case <-time.After(10 * time.Second):
		t.Fatal("the leftover of this install was not stopped")
	}
	if !isRunning(otherExited) {
		t.Error("the service of an install with a similar directory name was stopped")
	}
}
