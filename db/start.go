package db

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"

	"github.com/turbot/steampipe/cmdconfig"

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
	// NetworkListenType :: StartListenType - bind to all known interfaces
	NetworkListenType StartListenType = "network"
	// LocalListenType :: StartListenType - bind to localhost only
	LocalListenType = "local"
)

const (
	// ServiceInvoker :: Invoker - when invoked by `service start`
	ServiceInvoker Invoker = "service"
	// QueryInvoker :: Invoker - when invoked by `query`
	QueryInvoker = "query"
	// InstallerInvoker :: Invoker - when invoked by the `installer`
	InstallerInvoker = "installer"
	// PluginInvoker :: Invoker - when invoked by the `pluginmanager`
	PluginInvoker = "plugin"
)

// IsValid :: validator for StartListenType known values
func (slt StartListenType) IsValid() error {
	switch slt {
	case NetworkListenType, LocalListenType:
		return nil
	}
	return fmt.Errorf("Invalid listen type. Can be one of '%v' or '%v'", NetworkListenType, LocalListenType)
}

// IsValid :: validator for Invoker known values
func (slt Invoker) IsValid() error {
	switch slt {
	case ServiceInvoker, QueryInvoker, InstallerInvoker, PluginInvoker:
		return nil
	}
	return fmt.Errorf("Invalid invoker. Can be one of '%v', '%v', '%v' or '%v'", ServiceInvoker, QueryInvoker, InstallerInvoker, PluginInvoker)
}

// StartDB :: start the database is not already running
func StartDB(port int, listen StartListenType, invoker Invoker) (StartResult, error) {
	info, err := loadRunningInstanceInfo()

	if err != nil {
		return ServiceFailedToStart, err
	}

	if info != nil {
		processRunning, err := pidExists(info.Pid)
		if err != nil {
			return ServiceFailedToStart, err
		}
		if processRunning {
			return ServiceAlreadyRunning, nil
		}
	}

	// we need to start the process

	// remove the stale info file, ignoring errors - will overwrite anyway
	_ = removeRunningInstanceInfo()

	listenAddresses := "localhost"

	if listen == NetworkListenType {
		listenAddresses = "*"
	}

	if !isPortBindable(port) {
		return ServiceFailedToStart, fmt.Errorf("Cannot listen on %d. Are you sure that the interface is free?", port)
	}

	postgresCmd := exec.Command(
		getPostgresBinaryExecutablePath(),
		// by this time, we are sure that the port if free to listen to
		"-p", fmt.Sprint(port),
		"-c", fmt.Sprintf("listen_addresses=\"%s\"", listenAddresses),
		"-c", fmt.Sprintf("application_name=\"%s\"", constants.APPNAME),
		"-c", fmt.Sprintf("cluster_name=\"%s\"", constants.APPNAME),
		"-c", "autovacuum=off",
		"-c", "bgwriter_lru_maxpages=0",
		"-c", "effective-cache-size=64kB",
		"-c", "fsync=off",
		"-c", "full_page_writes=off",
		"-c", "maintenance-work-mem=1024kB",
		"-c", "password_encryption=scram-sha-256",
		"-c", "random-page-cost=0.01",
		"-c", "seq-page-cost=0.01",
		"-c", "shared-buffers=128kB",
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
		"-c", fmt.Sprintf("log_directory=%s", getDatabaseLogDirectory()),
		"-c", fmt.Sprintf("log_filename=%s", "postgresql-%Y-%m-%d.log"),

		// Data Directory
		"-D", getDataLocation())

	postgresCmd.Env = os.Environ()

	postgresCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:    true,
		Foreground: false,
	}

	err = postgresCmd.Start()

	if err != nil {
		return ServiceFailedToStart, err
	}

	runningInfo := new(RunningDBInstanceInfo)
	runningInfo.Pid = postgresCmd.Process.Pid
	runningInfo.Port = port
	runningInfo.User = constants.DatabaseSuperUser
	runningInfo.Database = "postgres"
	runningInfo.ListenType = listen
	runningInfo.Invoker = invoker

	runningInfo.Listen = []string{"localhost", "127.0.0.1"}
	if listen == NetworkListenType {
		addrs, _ := localAddresses()
		runningInfo.Listen = append(addrs, runningInfo.Listen...)
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
		if err = refreshConnections(client); err != nil {
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
	allProcesses, _ := process.Processes()
	for _, p := range allProcesses {
		cmdLine, _ := p.CmdlineSlice()
		if len(cmdLine) < 1 {
			continue
		}
		executable := cmdLine[0]

		// this is a steampipe postgres, kill it along with it's children
		if executable == getPostgresBinaryExecutablePath() {
			killProcessTree(p)
			return true
		}
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
