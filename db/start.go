package db

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/utils"

	"github.com/shirou/gopsutil/process"
	"github.com/turbot/steampipe/constants"
)

// StartResult :: pseudoEnum for outcomes of Start
type StartResult int

// StartListenType :: pseudoEnum of network binding for postgres
type StartListenType string

// Invoker :: pseudoEnum for what starts the service
type Invoker string

const (
	// ServiceStarted :: StartResult - Service was started
	ServiceStarted StartResult = iota
	// ServiceAlreadyRunning :: StartResult - Service was already running
	ServiceAlreadyRunning
	// ServiceFailedToStart :: StartResult - Could not start service
	ServiceFailedToStart
)

const (
	// ListenTypeNetwork :: StartListenType - bind to all known interfaces
	ListenTypeNetwork StartListenType = "network"
	// ListenTypeLocal :: StartListenType - bind to localhost only
	ListenTypeLocal = "local"
)

const (
	// InvokerService :: Invoker - when invoked by `service start`
	InvokerService Invoker = "service"
	// InvokerQuery :: Invoker - when invoked by `query`
	InvokerQuery = "query"
	// InvokerInstaller :: Invoker - when invoked by the `installer`
	InvokerInstaller = "installer"
	// InvokerPlugin :: Invoker - when invoked by the `pluginmanager`
	InvokerPlugin = "plugin"
)

// IsValid :: validator for StartListenType known values
func (slt StartListenType) IsValid() error {
	switch slt {
	case ListenTypeNetwork, ListenTypeLocal:
		return nil
	}
	return fmt.Errorf("Invalid listen type. Can be one of '%v' or '%v'", ListenTypeNetwork, ListenTypeLocal)
}

// IsValid :: validator for Invoker known values
func (slt Invoker) IsValid() error {
	switch slt {
	case InvokerService, InvokerQuery, InvokerInstaller, InvokerPlugin:
		return nil
	}
	return fmt.Errorf("Invalid invoker. Can be one of '%v', '%v', '%v' or '%v'", InvokerService, InvokerQuery, InvokerInstaller, InvokerPlugin)
}

// StartDB :: start the database is not already running
func StartDB(port int, listen StartListenType, invoker Invoker) (StartResult, error) {
	info, err := GetStatus()

	if err != nil {
		return ServiceFailedToStart, err
	}

	if info != nil {
		processRunning, err := pidExists(info.Pid)
		if err != nil {
			return ServiceFailedToStart, err
		}
		if processRunning {
			if info.Invoker == InvokerQuery {
				return ServiceAlreadyRunning, fmt.Errorf("You have a %s session open. Close this session before running %s.\nTo kill existing sessions, run %s", constants.Bold("steampipe query"), constants.Bold("steampipe service stop"), constants.Bold("steampipe service stop --force"))
			}
			return ServiceAlreadyRunning, nil
		}
	}

	// we need to start the process

	// remove the stale info file, ignoring errors - will overwrite anyway
	_ = removeRunningInstanceInfo()

	listenAddresses := "localhost"

	if listen == ListenTypeNetwork {
		listenAddresses = "*"
	}

	if !isPortBindable(port) {
		return ServiceFailedToStart, fmt.Errorf("Cannot listen on %d. Are you sure that the interface is free?", port)
	}

	checkedPreviousInstances := make(chan bool, 1)
	s := utils.StartSpinnerAfterDelay("Checking for running instances", constants.SpinnerShowTimeout, checkedPreviousInstances)
	previousProcess := findSteampipePostgresInstance()
	checkedPreviousInstances <- true
	utils.StopSpinner(s)
	if previousProcess != nil {
		return ServiceFailedToStart, fmt.Errorf("Another Steampipe service is already running. Use %s to kill all running instances before continuing.", constants.Bold("steampipe service stop --force"))
	}

	postgresCmd := exec.Command(
		getPostgresBinaryExecutablePath(),
		// by this time, we are sure that the port if free to listen to
		"-p", fmt.Sprint(port),
		"-c", fmt.Sprintf("listen_addresses=\"%s\"", listenAddresses),
		// NOTE: If quoted, the application name includes the quotes. Worried about
		// having spaces in the APPNAME, but leaving it unquoted since currently
		// the APPNAME is hardcoded to be steampipe.
		"-c", fmt.Sprintf("application_name=%s", constants.APPNAME),
		"-c", fmt.Sprintf("cluster_name=%s", constants.APPNAME),
		"-c", "autovacuum=off",
		"-c", "bgwriter_lru_maxpages=0",
		"-c", "effective-cache-size=64kB",
		"-c", "fsync=off",
		"-c", "full_page_writes=off",
		"-c", "maintenance-work-mem=1024kB",
		"-c", "password_encryption=scram-sha-256",
		"-c", "random-page-cost=0.01",
		"-c", "seq-page-cost=0.01",
		// "-c", "shared-buffers=128kB",
		"-c", "synchronous_commit=off",
		"-c", "temp-buffers=800kB",
		"-c", "timezone=UTC",
		"-c", "track_activities=off",
		"-c", "track_counts=off",
		"-c", "wal-buffers=32kB",
		"-c", "work-mem=64kB",
		"-c", "jit=off",

		// postgres log collection
		"-c", "log_statement=all",
		"-c", "log_min_duration_statement=2000",
		"-c", "logging_collector=on",
		"-c", "log_min_error_statement=error",
		"-c", fmt.Sprintf("log_directory=%s", constants.LogDir()),
		"-c", fmt.Sprintf("log_filename=%s", "postgresql-%Y-%m-%d.log"),

		// Data Directory
		"-D", getDataLocation())

	log.Println("[TRACE] postgres start command: ", postgresCmd)

	postgresCmd.Env = os.Environ()

	postgresCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:    true,
		Foreground: false,
	}

	err = postgresCmd.Start()

	if err != nil {
		return ServiceFailedToStart, err
	}

	// get the password file
	passwords, err := getPasswords()

	runningInfo := new(RunningDBInstanceInfo)
	runningInfo.Pid = postgresCmd.Process.Pid
	runningInfo.Port = port
	runningInfo.User = constants.DatabaseUser
	runningInfo.Password = passwords.Steampipe
	runningInfo.Database = constants.DatabaseName
	runningInfo.ListenType = listen
	runningInfo.Invoker = invoker

	runningInfo.Listen = constants.DatabaseListenAddresses
	if listen == ListenTypeNetwork {
		addrs, _ := localAddresses()
		runningInfo.Listen = append(runningInfo.Listen, addrs...)
	}

	if err := postgresCmd.Process.Release(); err != nil {
		return ServiceStarted, err
	}

	if err := saveRunningInstanceInfo(runningInfo); err != nil {
		return ServiceStarted, err
	}

	client, err := GetClient(false)

	if err != nil {
		// if we got an error here, then there probably was a problem
		// starting up the process. this may be because of a stray
		// steampipe postgres running that we don't know of. try killing it
		if killPreviousInstanceIfAny() {
			// remove info file (if any)
			_ = removeRunningInstanceInfo()
			// try restarting
			return StartDB(port, listen, invoker)
		}

		// there was nothing to kill.
		// this is some other problem that we are not accounting for
		return ServiceFailedToStart, err
	}

	// refresh plugin connections - ensure db schemas are in sync with connection config
	// NOTE: refresh defaulyts to true but will be set to false if this service start command has been invoked by a query command
	if cmdconfig.Viper().GetBool("refresh") {
		if err = RefreshConnections(client); err != nil {
			return ServiceStarted, err
		}
		if err = refreshFunctions(client); err != nil {
			return ServiceStarted, err
		}
	}

	return ServiceStarted, nil
}

func isPortBindable(port int) bool {
	// resolve an address to 127.0.0.1 for the given port
	addrString := fmt.Sprintf("127.0.0.1:%d", port)
	addr, err := net.ResolveTCPAddr("tcp", addrString)
	if err != nil {
		return false
	}
	// check that the given port can be used
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return false
	}
	defer l.Close()
	return true
}

// kill all postgres processes that were started as part of steampipe (if any)
func killPreviousInstanceIfAny() bool {
	p := findSteampipePostgresInstance()
	if p != nil {
		killProcessTree(p)
		return true
	}
	return false
}

func findSteampipePostgresInstance() *process.Process {
	allProcesses, _ := process.Processes()
	for _, p := range allProcesses {
		cmdLine, _ := p.CmdlineSlice()
		if isSteampipePostgresProcess(cmdLine) {
			return p
		}
	}
	return nil
}

func isSteampipePostgresProcess(cmdline []string) bool {
	if len(cmdline) < 1 {
		return false
	}
	if strings.Contains(cmdline[0], "postgres") {
		// this is a postgres process
		return helpers.StringSliceContains(cmdline, fmt.Sprintf("application_name=%s", constants.APPNAME))
	}
	return false
}

func killProcessTree(p *process.Process) {
	children, _ := p.Children()
	for _, child := range children {
		killProcessTree(child)
	}
	p.Kill()
}

func localAddresses() ([]string, error) {
	addresses := []string{}
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPNet:
				isToInclude := v.IP.IsGlobalUnicast() && (v.IP.To4() != nil)
				if isToInclude {
					addresses = append(addresses, v.IP.String())
				}
			}

		}
	}

	return addresses, nil
}
