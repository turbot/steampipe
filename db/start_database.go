package db

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/process"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/display"
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
	// InvokerCheck :: Invoker - when invoked by `check`
	InvokerCheck = "check"
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
	case InvokerService, InvokerQuery, InvokerCheck, InvokerInstaller, InvokerPlugin:
		return nil
	}
	return fmt.Errorf("Invalid invoker. Can be one of '%v', '%v', '%v' or '%v'", InvokerService, InvokerQuery, InvokerInstaller, InvokerPlugin)
}

// StartDB :: start the database is not already running
func StartDB(port int, listen StartListenType, invoker Invoker) (startResult StartResult, err error) {
	defer func() {
		// if there was an error and we started the service, stop it again
		if err != nil {
			if startResult == ServiceStarted {
				StopDB(false, invoker)
			}
		}
	}()
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
		return ServiceFailedToStart, fmt.Errorf("Cannot listen on port %s. To start the service with a different port, use %s", constants.Bold(port), constants.Bold("--database-port <number>"))
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
		// If the shared buffers are too small then large tables in memory can create
		// "no unpinned buffers available" errors.
		// "-c", "shared-buffers=128kB",
		// If synchronous_commit=off then the setup process can fail because the
		// installation of the foreign server is not committed before the DB shutsdown.
		// Steampipe does very few commits in general, so leaving this on will have
		// very little impact on performance.
		// "-c", "synchronous_commit=off",
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
		"-c", fmt.Sprintf("log_filename=%s", "database-%Y-%m-%d.log"),

		// Data Directory
		"-D", getDataLocation())

	postgresCmd.Env = append(os.Environ(), fmt.Sprintf("STEAMPIPE_INSTALL_DIR=%s", constants.SteampipeDir))

	log.Println("[TRACE] postgres start command: ", postgresCmd.String())
	log.Println("[TRACE] postgres environment: ", postgresCmd.Env)

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

	// create a client
	// pass 'false' to disable auto refreshing connections
	//- we will explicitly refresh connections after ensuring the steampipe server exists
	client, err := NewClient(false)
	if err != nil {
		return ServiceFailedToStart, handleStartFailure(err)
	}
	defer func() {
		client.Close()
	}()

	err = ensureSteampipeServer()
	if err != nil {
		// there was a problem with the installation
		StopDB(true, invoker)
		return ServiceFailedToStart, err
	}

	// refresh plugin connections - ensure db schemas are in sync with connection config
	// NOTE: refresh defaults to true but will be set to false if this service start command has been invoked by a query command
	if cmdconfig.Viper().GetBool(constants.ArgRefresh) {
		if _, err = client.RefreshConnections(); err != nil {
			return ServiceStarted, err
		}
		if err = refreshFunctions(); err != nil {
			return ServiceStarted, err
		}
	}

	err = client.SetServiceSearchPath()
	return ServiceStarted, err
}

// ensures that the `steampipe` fdw server exists
// checks for it - (re)installs FDW and creates server if it doesn't
func ensureSteampipeServer() error {
	rootClient, err := createSteampipeRootDbClient()
	if err != nil {
		return err
	}
	defer rootClient.Close()
	out := rootClient.QueryRow("select srvname from pg_catalog.pg_foreign_server where srvname='steampipe'")
	var serverName string
	err = out.Scan(&serverName)

	if err != nil {
		return installSteampipeHub()
	}
	return nil
}

func handleStartFailure(err error) error {
	// if we got an error here, then there probably was a problem
	// starting up the process. this may be because of a stray
	// steampipe postgres running or another one from a different installation.
	checkedPreviousInstances := make(chan bool, 1)
	s := display.StartSpinnerAfterDelay("Checking for running instances...", constants.SpinnerShowTimeout, checkedPreviousInstances)
	otherProcess := findSteampipePostgresInstance()
	checkedPreviousInstances <- true
	display.StopSpinner(s)
	if otherProcess != nil {
		return fmt.Errorf("Another Steampipe service is already running. Use %s to kill all running instances before continuing.", constants.Bold("steampipe service stop --force"))
	}

	// there was nothing to kill.
	// this is some other problem that we are not accounting for
	return err
}

func isPortBindable(port int) bool {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	defer l.Close()
	return true
}

// kill all postgres processes that were started as part of steampipe (if any)
func killPreviousInstanceIfAny() bool {
	wasKilled := false
	for {
		p := findSteampipePostgresInstance()
		if p == nil {
			break
		}
		killProcessTree(p)
		wasKilled = true
	}
	return wasKilled
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
