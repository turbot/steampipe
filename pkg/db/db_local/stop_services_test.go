package db_local

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

func TestShutdownService_LeavesAServiceAnOlderVersionStartedRunning(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in service is sleep")
	}
	originalInstallDir := app_specific.InstallDir
	t.Cleanup(func() { app_specific.InstallDir = originalInstallDir })
	app_specific.InstallDir = filepath.Join(shortTempDir(t), ".steampipe")

	// a stand-in for a postgres that an older version started for a query: it has no socket in the data
	// directory, so the connected clients cannot be counted
	service := exec.Command("sleep", "60")
	if err := service.Start(); err != nil {
		t.Fatal(err)
	}
	exited := make(chan struct{})
	go func() {
		_ = service.Wait()
		close(exited)
	}()
	t.Cleanup(func() { _ = service.Process.Kill() })
	info := newRunningDBInstanceInfo(service, []string{"127.0.0.1"}, 9193, "steampipe", "password", constants.InvokerQuery)
	if err := info.Save(); err != nil {
		t.Fatal(err)
	}

	ShutdownService(context.Background(), constants.InvokerQuery)

	select {
	case <-exited:
		t.Error("a service whose clients cannot be counted must not be stopped: another session may be using it")
	case <-time.After(500 * time.Millisecond):
	}
}
