package db

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/utils"
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
	return fmt.Errorf("Invalid invoker. Can be one of '%v', '%v', '%v', '%v' or '%v'", InvokerService, InvokerQuery, InvokerInstaller, InvokerPlugin, InvokerCheck)
}

// StartDB :: start the database is not already running
func StartDB(port int, listen StartListenType, invoker Invoker, refreshConnections bool) (startResult StartResult, err error) {
	utils.LogTime("db.StartDB start")
	defer utils.LogTime("db.StartDB end")

	var client *Client

	defer func() {
		// if there was an error and we started the service, stop it again
		if err != nil {
			if startResult == ServiceStarted {
				StopDB(false, invoker, nil)
			}
		}

		if client != nil {
			client.Close()
		}
	}()
	info, err := GetStatus()

	if err != nil {
		return ServiceFailedToStart, err
	}

	if info != nil {
		// check whether the stated PID actually exists
		processRunning, err := PidExists(info.Pid)
		if err != nil {
			return ServiceFailedToStart, err
		}

		// Process with declared PID exists.
		// Check if the service was started by another `service` command
		// if not, throw an error.
		if processRunning {
			if info.Invoker != InvokerService {
				return ServiceAlreadyRunning, fmt.Errorf("You have a %s session open. Close this session before running %s.\nTo force kill all existing sessions, run %s", constants.Bold(fmt.Sprintf("steampipe %s", info.Invoker)), constants.Bold("steampipe service start"), constants.Bold("steampipe service stop --force"))
			}
			return ServiceAlreadyRunning, nil
		}
	}

	// we need to start the process

	// remove the stale info file, ignoring errors - will overwrite anyway
	_ = removeRunningInstanceInfo()

	// Generate the certificate if it fails then set the ssl to off
	if err := generateSelfSignedCertificate(); err != nil {
		utils.ShowWarning("self signed certificate creation failed, connecting to the database without SSL")
	}

	if err := isPortBindable(port); err != nil {
		return ServiceFailedToStart, fmt.Errorf("cannot listen on port %d", constants.Bold(port))
	}

	utils.LogTime("postgresCmd start")
	err = startPostgresProcess(port, listen, invoker)
	if err != nil {
		return ServiceFailedToStart, err
	}
	utils.LogTime("postgresCmd end")

	// create a client
	// pass 'false' to disable auto refreshing connections
	//- we will explicitly refresh connections after ensuring the steampipe server exists
	if refreshConnections {
		client, err = NewClient()
		if err != nil {
			return ServiceFailedToStart, handleStartFailure(err)
		}
		refreshResult := client.RefreshConnectionAndSearchPaths()
		if refreshResult.Error != nil {
			return ServiceFailedToStart, handleStartFailure(refreshResult.Error)
		}
		// display any initialisation warnings
		refreshResult.ShowWarnings()
	}
	err = ensureSteampipeServer()
	if err != nil {
		// there was a problem with the installation
		return ServiceFailedToStart, err
	}

	err = ensureTempTablePermissions()
	if err != nil {
		// there was a problem with the installation
		return ServiceFailedToStart, err
	}

	return ServiceStarted, err
}

// startPostgresProcess starts the postgres process and writes out the state file
// after it is convinced that the process is started and is accepting connections
func startPostgresProcess(port int, listen StartListenType, invoker Invoker) error {
	utils.LogTime("startPostgresProcess start")
	defer utils.LogTime("startPostgresProcess end")

	listenAddresses := "localhost"

	if listen == ListenTypeNetwork {
		listenAddresses = "*"
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

		// If ssl is off  it doesnot matter what we pass in the ssl_cert_file and ssl_key_file
		// SSL will only get validated if the ssl is on
		"-c", fmt.Sprintf("ssl=%s", SslStatus()),
		"-c", fmt.Sprintf("ssl_cert_file=%s", filepath.Join(getDataLocation(), constants.ServerCert)),
		"-c", fmt.Sprintf("ssl_key_file=%s", filepath.Join(getDataLocation(), constants.ServerKey)),

		// Data Directory
		"-D", getDataLocation())

	postgresCmd.Env = append(os.Environ(), fmt.Sprintf("STEAMPIPE_INSTALL_DIR=%s", constants.SteampipeDir))

	//  Check if the /etc/ssl directory exist in os
	dirExist, _ := os.Stat(constants.SslConfDir)
	_, envVariableExist := os.LookupEnv("OPENSSL_CONF")

	// This is particularly required for debian:buster
	// https://github.com/kelaberetiv/TagUI/issues/787
	// For other os the env variable OPENSSL_CONF
	// does not matter so its safe to put
	// this in env variable
	// Tested in amazonlinux, debian:buster, ubuntu, mac
	if dirExist != nil && !envVariableExist {
		postgresCmd.Env = append(os.Environ(), fmt.Sprintf("OPENSSL_CONF=%s", constants.SslConfDir))
	}

	postgresCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:    true,
		Foreground: false,
	}

	err := postgresCmd.Start()
	if err != nil {
		return err
	}

	// get the password file
	passwords, _ := getPasswords()

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
	err = runningInfo.Save()
	if err != nil {
		postgresCmd.Process.Kill()
		return err
	}

	err = postgresCmd.Process.Release()
	if err != nil {
		postgresCmd.Process.Kill()
		return err
	}

	connection, err := createDbClient("postgres", constants.DatabaseSuperUser)
	if err != nil {
		postgresCmd.Process.Kill()
		return err
	}
	connection.Close()

	return nil
}

// ensures that the `steampipe` fdw server exists
// checks for it - (re)install FDW and creates server if it doesn't
func ensureSteampipeServer() error {
	rootClient, err := createRootDbClient()
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

// ensures that the `steampipe_users` role has permissions to work with temporary tables
// this is done during database installation, but we need to migrate current installations
func ensureTempTablePermissions() error {
	rootClient, err := createRootDbClient()
	if err != nil {
		return err
	}
	defer rootClient.Close()
	_, err = rootClient.Exec("grant temporary on database steampipe to steampipe_users")
	if err != nil {
		return err
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
	close(checkedPreviousInstances)
	display.StopSpinner(s)
	if otherProcess != nil {
		return fmt.Errorf("Another Steampipe service is already running. Use %s to kill all running instances before continuing.", constants.Bold("steampipe service stop --force"))
	}

	// there was nothing to kill.
	// this is some other problem that we are not accounting for
	return err
}

func isPortBindable(port int) error {
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return err
	}
	defer l.Close()
	return nil
}

// kill all postgres processes that were started as part of steampipe (if any)
func killInstanceIfAny() bool {
	processes, err := FindAllSteampipePostgresInstances()
	if err != nil {
		return false
	}
	wg := sync.WaitGroup{}
	for _, process := range processes {
		wg.Add(1)
		go func(p *psutils.Process) {
			doThreeStepPostgresExit(p)
			wg.Done()
		}(process)
	}
	wg.Wait()
	return len(processes) > 0
}

func FindAllSteampipePostgresInstances() ([]*psutils.Process, error) {
	instances := []*psutils.Process{}
	allProcesses, err := psutils.Processes()
	if err != nil {
		return nil, err
	}
	for _, p := range allProcesses {
		if cmdLine, err := p.CmdlineSlice(); err == nil {
			if isSteampipePostgresProcess(cmdLine) {
				instances = append(instances, p)
			}
		} else {
			return nil, err
		}
	}
	return instances, nil
}

func findSteampipePostgresInstance() *psutils.Process {
	allProcesses, _ := psutils.Processes()
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
		// this is a postgres process - but is it a steampipe service?
		return helpers.StringSliceContains(cmdline, fmt.Sprintf("application_name=%s", constants.APPNAME))
	}
	return false
}

func killProcessTree(p *psutils.Process) error {
	// find it's children
	children, err := p.Children()
	if err != nil {
		return err
	}
	for _, child := range children {
		// and kill them first
		killProcessTree(child)
	}
	p.Kill()
	return nil
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
