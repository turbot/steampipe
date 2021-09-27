package local_db

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

// StartResult :: pseudoEnum for outcomes of Start
type StartResult int

// StartListenType is a pseudoEnum of network binding for postgres
type StartListenType string

const (
	// ServiceStarted :: StartResult - Service was started
	// start from 10 to prevent confusion with int zero-value
	ServiceStarted StartResult = iota + 1
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

// IsValid is a validator for StartListenType known values
func (slt StartListenType) IsValid() error {
	switch slt {
	case ListenTypeNetwork, ListenTypeLocal:
		return nil
	}
	return fmt.Errorf("Invalid listen type. Can be one of '%v' or '%v'", ListenTypeNetwork, ListenTypeLocal)
}

// StartDB starts the database if not already running
func StartDB(port int, listen StartListenType, invoker constants.Invoker) (startResult StartResult, err error) {
	utils.LogTime("db.StartDB start")
	defer utils.LogTime("db.StartDB end")

	var client *LocalClient

	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
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
			if info.Invoker != constants.InvokerService {
				return ServiceAlreadyRunning, fmt.Errorf("You have a %s session open. Close this session before running %s.\nTo force kill all existing sessions, run %s", constants.Bold(fmt.Sprintf("steampipe %s", info.Invoker)), constants.Bold("steampipe service start"), constants.Bold("steampipe service stop --force"))
			}
			return ServiceAlreadyRunning, nil
		}
	}

	// we need to start the process

	// remove the stale info file, ignoring errors - will overwrite anyway
	_ = removeRunningInstanceInfo()

	if err := utils.EnsureDirectoryPermission(getDataLocation()); err != nil {
		return ServiceFailedToStart, fmt.Errorf("%s does not have the necessary permissions to start the service", getDataLocation())
	}

	// Generate the certificate if it fails then set the ssl to off
	if err := generateSelfSignedCertificate(); err != nil {
		utils.ShowWarning("self signed certificate creation failed, connecting to the database without SSL")
	}

	if err := isPortBindable(port); err != nil {
		return ServiceFailedToStart, fmt.Errorf("cannot listen on port %d", constants.Bold(port))
	}

	if err := migrateLegacyPasswordFile(); err != nil {
		return ServiceFailedToStart, err
	}

	utils.LogTime("postgresCmd start")
	err = startPostgresProcessAndSetPassword(port, listen, invoker)
	if err != nil {
		return ServiceFailedToStart, err
	}
	utils.LogTime("postgresCmd end")

	err = ensureSteampipeServer()
	if err != nil {
		// there was a problem with the installation
		return ServiceFailedToStart, err
	}

	err = ensureTempTablePermissions()
	if err != nil {
		return ServiceFailedToStart, err
	}
	// ensure the db contains command schema
	err = ensureCommandSchema()
	if err != nil {
		return ServiceFailedToStart, err
	}

	return ServiceStarted, err
}

// startPostgresProcessAndSetPassword starts the postgres process and writes out the state file
// after it is convinced that the process is started and is accepting connections
func startPostgresProcessAndSetPassword(port int, listen StartListenType, invoker constants.Invoker) (e error) {
	utils.LogTime("startPostgresProcess start")
	defer utils.LogTime("startPostgresProcess end")

	defer func() {
		if e != nil {
			// remove the state file if we are going back with an error
			removeRunningInstanceInfo()
		}
	}()

	listenAddresses := "localhost"

	if listen == ListenTypeNetwork {
		listenAddresses = "*"
	}

	if err := writePGConf(); err != nil {
		return err
	}

	postgresCmd := createCmd(port, listenAddresses)
	setupLogCollection(postgresCmd)
	err := postgresCmd.Start()
	if err != nil {
		return err
	}

	// get the password from the password file
	password, err := readPasswordFile()
	if err != nil {
		return err
	}

	// if a password was set through the `STEAMPIPE_DATABASE_PASSWORD` environment variable
	// or through the `--database-password` cmdline flag, then use that for this session
	// instead of the default one
	if viper.IsSet(constants.ArgServicePassword) {
		password = viper.GetString(constants.ArgServicePassword)
	}

	err = createRunningInfo(postgresCmd, port, password, listen, invoker)
	if err != nil {
		postgresCmd.Process.Kill()
		return err
	}

	err = postgresCmd.Process.Release()
	if err != nil {
		postgresCmd.Process.Kill()
		return err
	}

	// set the password on the database
	// we can't do this during installation, since the 'steampipe` user isn't setup yet
	if invoker != constants.InvokerInstaller {
		err = setupServicePassword(invoker, password)
		if err != nil {
			postgresCmd.Process.Kill()
			return err
		}
	}
	return nil
}

func writePGConf() error {
	// Apply default settings in conf files
	err := ioutil.WriteFile(getPostgresqlConfLocation(), []byte(constants.PostgresqlConfContent), 0600)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(getSteampipeConfLocation(), []byte(constants.SteampipeConfContent), 0600)
	if err != nil {
		return err
	}

	// create the postgresql.conf.d location, don't fail if it errors
	err = os.MkdirAll(getPostgresqlConfDLocation(), 0700)
	if err != nil {
		return err
	}
	return nil
}

func createCmd(port int, listenAddresses string) *exec.Cmd {
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

		// log directory
		"-c", fmt.Sprintf("log_directory=%s", constants.LogDir()),

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

	return postgresCmd
}

func setupLogCollection(cmd *exec.Cmd) {
	// create a channel with a big buffer, so that it doesn't choke
	logChannel := make(chan string, 1000)
	stopListenFn, err := setupLogCollector(cmd, logChannel)
	if err == nil {
		defer func() {
			stopListenFn()
		}()
		go traceoutServiceLogs(logChannel)
	} else {
		// this is a convenience and therefore, we shouldn't error out if we
		// are not able to capture the logs.
		// instead, log to TRACE that we couldn't and continue
		log.Println("[TRACE] Warning: Could not attach to service logs")
	}
}

func createRunningInfo(cmd *exec.Cmd, port int, password string, listen StartListenType, invoker constants.Invoker) error {
	runningInfo := new(RunningDBInstanceInfo)
	runningInfo.Pid = cmd.Process.Pid
	runningInfo.Port = port
	runningInfo.User = constants.DatabaseUser
	runningInfo.Password = password
	runningInfo.Database = constants.DatabaseName
	runningInfo.ListenType = listen
	runningInfo.Invoker = invoker
	runningInfo.Listen = constants.DatabaseListenAddresses

	if listen == ListenTypeNetwork {
		addrs, _ := localAddresses()
		runningInfo.Listen = append(runningInfo.Listen, addrs...)
	}
	err := runningInfo.Save()
	if err != nil {
		cmd.Process.Kill()
		return err
	}

	err = cmd.Process.Release()
	if err != nil {
		cmd.Process.Kill()
		return err
	}

	connection, err := createDbClient("postgres", constants.DatabaseSuperUser)
	if err != nil {
		cmd.Process.Kill()
		return err
	}
	connection.Close()

	return nil
}

func traceoutServiceLogs(logChannel chan string) {
	for logLine := range logChannel {
		log.Printf("[TRACE] SERVICE: %s\n", logLine)
		if strings.Contains(logLine, "Future log output will appear in") {
			break
		}
	}
}

func setupServicePassword(invoker constants.Invoker, password string) error {
	connection, err := createRootDbClient()
	if err != nil {
		return err
	}
	defer connection.Close()

	_, err = connection.Exec(fmt.Sprintf(`alter user steampipe with password '%s'`, password))
	return err
}

func setupLogCollector(postgresCmd *exec.Cmd, publishChannel chan string) (func(), error) {
	stdoutPipe, err := postgresCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderrPipe, err := postgresCmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	closeFunction := func() {
		stdoutPipe.Close()
		stderrPipe.Close()

		// always close from the sender
		close(publishChannel)
	}
	stdoutScanner := bufio.NewScanner(stdoutPipe)
	stderrScanner := bufio.NewScanner(stderrPipe)

	stdoutScanner.Split(bufio.ScanLines)
	stderrScanner.Split(bufio.ScanLines)

	go func() {
		for stdoutScanner.Scan() {
			line := stdoutScanner.Text()
			if len(line) > 0 {
				publishChannel <- line
			}
		}
	}()

	go func() {
		for stderrScanner.Scan() {
			line := stderrScanner.Text()
			if len(line) > 0 {
				publishChannel <- line
			}
		}
	}()

	return closeFunction, nil
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
	// if there is an error, we need to reinstall the foreign server
	if err != nil {
		return installForeignServer()
	}
	return nil
}

// create the command schema and grant insert permission
func ensureCommandSchema() error {
	if _, err := executeSqlAsRoot(updateConnectionQuery(constants.CommandSchema, constants.CommandSchema)...); err != nil {
		return err
	}
	_, err := executeSqlAsRoot(fmt.Sprintf("grant insert on %s.%s to steampipe_users;", constants.CommandSchema, constants.CacheCommandTable))

	return err
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
