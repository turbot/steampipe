package dashboardserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/shirou/gopsutil/process"
	"github.com/turbot/steampipe-plugin-sdk/v3/logging"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/dashboard/dashboardassets"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/utils"
)

type DashboardServiceState struct {
	Pid        int
	Port       int
	ListenType string
	Listen     []string
}

func GetDashboardServiceState() (*DashboardServiceState, error) {
	state := &DashboardServiceState{}
	fileContent, err := os.ReadFile(filepaths.DashboardServiceStateFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	err = json.Unmarshal(fileContent, state)
	if err != nil {
		return nil, err
	}
	pidExists, err := utils.PidExists(state.Pid)
	if err != nil {
		return nil, err
	}
	if !pidExists {
		return nil, os.Remove(filepaths.DashboardServiceStateFilePath())
	}
	return state, nil
}

func StopDashboardService(ctx context.Context) error {
	state, err := GetDashboardServiceState()
	if err != nil {
		return err
	}
	if state == nil {
		return nil
	}
	pidExists, err := utils.PidExists(state.Pid)
	if err != nil {
		return err
	}
	if !pidExists {
		return os.Remove(filepaths.DashboardServiceStateFilePath())
	}
	process, err := process.NewProcessWithContext(ctx, int32(state.Pid))
	if err != nil {
		return err
	}
	err = process.SendSignalWithContext(ctx, syscall.SIGINT)
	if err != nil {
		return err
	}
	return os.Remove(filepaths.DashboardServiceStateFilePath())
}

func RunForService(ctx context.Context, serverListen ListenType, serverPort ListenPort) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}

	err = dashboardassets.Ensure(ctx)
	if err != nil {
		return err
	}

	utils.FailOnError(serverPort.IsValid())
	utils.FailOnError(serverListen.IsValid())

	cmd := exec.Command(
		self,
		"dashboard",
		fmt.Sprintf("--%s=%s", constants.ArgDashboardListen, string(serverListen)),
		fmt.Sprintf("--%s=%d", constants.ArgDashboardPort, serverPort),
		fmt.Sprintf("--%s=%t", constants.ArgDashboardClient, false),
	)

	cmd.Env = append(os.Environ(), fmt.Sprintf("STEAMPIPE_INSTALL_DIR=%s", filepaths.SteampipeDir))

	// set group pgid attributes on the command to ensure the process is not shutdown when its parent terminates
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:    true,
		Foreground: false,
	}

	logger := setupDashboardServerLogSink()
	writer := logger.StandardWriter(&hclog.StandardLoggerOptions{ForceLevel: hclog.Trace})
	cmd.Stdout = writer
	cmd.Stderr = writer

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = waitForDashboardServerStartup(ctx, int(serverPort))
	if err != nil {
		return err
	}

	state := &DashboardServiceState{
		Pid:        cmd.Process.Pid,
		Port:       int(serverPort),
		ListenType: string(serverListen),
		Listen:     constants.DatabaseListenAddresses,
	}

	if serverListen == ListenTypeNetwork {
		addrs, _ := utils.LocalAddresses()
		state.Listen = append(state.Listen, addrs...)
	}

	stateBytes, err := json.Marshal(state)
	if err != nil {
		cmd.Process.Signal(syscall.SIGINT)
		return err
	}
	err = os.WriteFile(filepaths.DashboardServiceStateFilePath(), stateBytes, 0666)
	if err != nil {
		cmd.Process.Signal(syscall.SIGINT)
		return err
	}

	return nil
}

func waitForDashboardServerStartup(ctx context.Context, serverPort int) error {
	utils.LogTime("db.waitForConnection start")
	defer utils.LogTime("db.waitForConnection end")

	pingTimer := time.NewTicker(10 * time.Millisecond)
	timeoutAt := time.After(5 * time.Second)
	defer pingTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-pingTimer.C:
			_, err := http.Get(fmt.Sprintf("http://localhost:%d", serverPort))
			if err == nil {
				return err
			}
		case <-timeoutAt:
			return fmt.Errorf("dashboard server startup failed")
		}
	}
}

func setupDashboardServerLogSink() hclog.Logger {
	logName := fmt.Sprintf("dashboard-%s.log", time.Now().Format("2006-01-02"))
	logPath := filepath.Join(filepaths.EnsureLogDir(), logName)
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("failed to open plugin manager log file: %s\n", err.Error())
		os.Exit(3)
	}
	logger := logging.NewLogger(&hclog.LoggerOptions{
		Output:     f,
		TimeFn:     func() time.Time { return time.Now().UTC() },
		TimeFormat: "2006-01-02 15:04:05.000 UTC",
	})
	return logger
}
