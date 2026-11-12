package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/turbot/steampipe/v2/pkg/db/db_local"
)

func TestParseServiceCmdline_EmptyCmdline(t *testing.T) {
	for _, cmdLine := range [][]string{nil, {}} {
		_, _, _, _, ok := parseServiceCmdline(42, cmdLine)
		if ok {
			t.Errorf("expected ok=false for empty command line %#v", cmdLine)
		}
	}
}

func TestParseServiceCmdline_TrailingPortFlag(t *testing.T) {
	cmdLine := []string{"/home/u/.steampipe/db/14.19.0/postgres/bin/postgres", "-c", "listen_addresses=localhost", "-p"}
	pid, _, port, listen, ok := parseServiceCmdline(42, cmdLine)
	if !ok {
		t.Fatal("expected ok=true for a readable command line")
	}
	if pid != "42" {
		t.Errorf("pid = %q, want 42", pid)
	}
	if port != "" {
		t.Errorf("port = %q, want empty", port)
	}
	if listen != db_local.ListenTypeLocal {
		t.Errorf("listen = %q, want %q", listen, db_local.ListenTypeLocal)
	}
}

func TestParseServiceCmdline_FullCmdline(t *testing.T) {
	cmdLine := []string{"/x/postgres", "-p", "9193", "-c", "listen_addresses=*"}
	_, _, port, listen, ok := parseServiceCmdline(7, cmdLine)
	if !ok || port != "9193" || listen != db_local.ListenTypeNetwork {
		t.Errorf("got port=%q listen=%q ok=%v", port, listen, ok)
	}
}

func exitedProcess(t *testing.T) *psutils.Process {
	t.Helper()
	c := exec.Command("sleep", "30")
	if err := c.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = c.Process.Kill()
		_ = c.Wait()
	})
	p, err := psutils.NewProcess(int32(c.Process.Pid)) //nolint:gosec // a pid fits in int32
	if err != nil {
		t.Fatal(err)
	}
	_ = c.Process.Kill()
	_ = c.Wait()
	return p
}

func TestGetServiceProcessDetails_ExitedProcess(t *testing.T) {
	_, _, _, _, ok := getServiceProcessDetails(exitedProcess(t))
	if ok {
		t.Error("expected ok=false for an exited process")
	}
}

func TestBuildServiceRows_KeepsProcessesAfterAnExitedOne(t *testing.T) {
	live, err := psutils.NewProcess(int32(os.Getpid())) //nolint:gosec // a pid fits in int32
	if err != nil {
		t.Fatal(err)
	}
	rows := buildServiceRows([]*psutils.Process{exitedProcess(t), live})
	if len(rows) != 1 || rows[0][0] != fmt.Sprintf("%d", live.Pid) {
		t.Errorf("rows = %v, want a single row for pid %d", rows, live.Pid)
	}
}

func TestBuildServiceRows_AllProcessesExited(t *testing.T) {
	rows := buildServiceRows([]*psutils.Process{exitedProcess(t)})
	if len(rows) != 0 {
		t.Errorf("rows = %v, want none", rows)
	}
}
