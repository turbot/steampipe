package dashboardserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/process"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardassets"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/utils"
)

type ServiceState string

const (
	ServiceStateRunning       ServiceState = "running"
	ServiceStateError         ServiceState = "error"
	ServiceStateStructVersion              = 20220411
)

type DashboardServiceState struct {
	State         ServiceState `json:"state"`
	Error         string       `json:"error"`
	Pid           int          `json:"pid"`
	Port          int          `json:"port"`
	ListenType    string       `json:"listen_type"`
	Listen        []string     `json:"listen"`
	StructVersion int64        `json:"struct_version"`
}

func loadServiceStateFile() (*DashboardServiceState, error) {
	state := &DashboardServiceState{}
	stateBytes, err := os.ReadFile(filepaths.DashboardServiceStateFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	err = json.Unmarshal(stateBytes, state)
	return state, err
}

func (s *DashboardServiceState) Save() error {
	// set struct version
	s.StructVersion = ServiceStateStructVersion

	versionFilePath := filepaths.DashboardServiceStateFilePath()
	return s.write(versionFilePath)
}

func (s *DashboardServiceState) write(path string) error {
	versionFileJSON, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		log.Println("Error while writing version file", err)
		return err
	}
	return os.WriteFile(path, versionFileJSON, 0644)
}

func GetDashboardServiceState() (*DashboardServiceState, error) {
	state, err := loadServiceStateFile()
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, nil
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
		return nil
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

// RunForService spanws an execution of the 'steampipe dashboard' command.
// It is used when starting/restarting the steampipe service with the --dashboard flag set
func RunForService(ctx context.Context, serverListen ListenType, serverPort ListenPort) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}

	// remove the state file (if any)
	os.Remove(filepaths.DashboardServiceStateFilePath())

	err = dashboardassets.Ensure(ctx)
	if err != nil {
		return err
	}

	error_helpers.FailOnError(serverPort.IsValid())
	error_helpers.FailOnError(serverListen.IsValid())

	// NOTE: args must be specified <arg>=<arg val>, as each entry in this array is a separate arg passed to cobra
	args := []string{
		"dashboard",
		fmt.Sprintf("--%s=%s", constants.ArgDashboardListen, string(serverListen)),
		fmt.Sprintf("--%s=%d", constants.ArgDashboardPort, serverPort),
		fmt.Sprintf("--%s=%s", constants.ArgInstallDir, filepaths.SteampipeDir),
		fmt.Sprintf("--%s=%s", constants.ArgModLocation, viper.GetString(constants.ArgModLocation)),
		fmt.Sprintf("--%s=true", constants.ArgServiceMode),
		fmt.Sprintf("--%s=false", constants.ArgInput),
	}

	for _, variableArg := range viper.GetStringSlice(constants.ArgVariable) {
		args = append(args, fmt.Sprintf("--%s=%s", constants.ArgVariable, variableArg))
	}

	for _, varFile := range viper.GetStringSlice(constants.ArgVarFile) {
		args = append(args, fmt.Sprintf("--%s=%s", constants.ArgVarFile, varFile))
	}
	cmd := exec.Command(
		self,
		args...,
	)
	cmd.Env = os.Environ()

	// set group pgid attributes on the command to ensure the process is not shutdown when its parent terminates
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:    true,
		Foreground: false,
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	return waitForDashboardService(ctx)
}

// when started as a service, 'steampipe dashboard' always writes a
// state file in 'internal' with the outcome - even on failures
// this function polls for the file and loads up the error, if any
func waitForDashboardService(ctx context.Context) error {
	utils.LogTime("db.waitForDashboardServerStartup start")
	defer utils.LogTime("db.waitForDashboardServerStartup end")

	pingTimer := time.NewTicker(constants.ServicePingInterval)
	timeoutAt := time.After(constants.DashboardServiceStartTimeout)
	defer pingTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-pingTimer.C:
			// poll for the state file.
			// when it comes up, return it
			state, err := loadServiceStateFile()
			if err != nil {
				if os.IsNotExist(err) {
					// if the file hasn't been generated yet, that means 'dashboard' is still booting up
					continue
				}
				// there was an unexpected error
				return err
			}

			if state == nil {
				// no state file yet
				continue
			}

			// check the state file for an error
			if len(state.Error) > 0 {
				// there was an error during start up
				// remove the state file, since we don't need it anymore
				os.Remove(filepaths.DashboardServiceStateFilePath())
				// and return the error from the state file
				return errors.New(state.Error)
			}

			// we loaded the state and there was no error
			return nil
		case <-timeoutAt:
			return fmt.Errorf("dashboard server startup timed out")
		}
	}
}

func WriteServiceStateFile(state *DashboardServiceState) error {
	stateBytes, err := json.MarshalIndent(state, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepaths.DashboardServiceStateFilePath(), stateBytes, 0666)
}
