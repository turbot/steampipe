package db_local

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/turbot/steampipe/plugin_manager"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/utils"
)

// StartResult is a pseudoEnum for outcomes of Start
type StartResult int

// StartListenType is a pseudoEnum of network binding for postgres
type StartListenType string

const (
	// start from 10 to prevent confusion with int zero-value
	ServiceStarted StartResult = iota + 1
	ServiceAlreadyRunning
	ServiceFailedToStart
)

const (
	// ListenTypeNetwork - bind to all known interfaces
	ListenTypeNetwork StartListenType = "network"
	// ListenTypeLocal - bind to localhost only
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
	log.Printf("[TRACE] StartDB invoker %s", invoker)
	utils.LogTime("db.StartDB start")
	defer utils.LogTime("db.StartDB end")
	var postgresCmd *exec.Cmd

	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
		// if there was an error and we started the service, stop it again
		if err != nil {
			if startResult == ServiceStarted {
				StopDB(false, invoker, nil)
			}
			// remove the state file if we are going back with an error
			removeRunningInstanceInfo()
			// we are going back with an error
			// if the process was started,
			if postgresCmd != nil && postgresCmd.Process != nil {
				// kill it
				postgresCmd.Process.Kill()
			}
		}
	}()

	// start the plugin manager
	log.Printf("[WARN] StartDB starting plugin manager")
	if err := plugin_manager.Start(); err != nil {
		return ServiceFailedToStart, err
	}
	log.Printf("[WARN] StartDB started plugin manager")

	// remove the stale info file, ignoring errors - will overwrite anyway
	_ = removeRunningInstanceInfo()

	if err := utils.EnsureDirectoryPermission(getDataLocation()); err != nil {
		return ServiceFailedToStart, fmt.Errorf("%s does not have the necessary permissions to start the service", getDataLocation())
	}

	// Generate the certificate if it fails then set the ssl to off
	if err := ensureSelfSignedCertificate(); err != nil {
		utils.ShowWarning("self signed certificate creation failed, connecting to the database without SSL")
	}

	if err := isPortBindable(port); err != nil {
		return ServiceFailedToStart, fmt.Errorf("cannot listen on port %d", constants.Bold(port))
	}

	if err := migrateLegacyPasswordFile(); err != nil {
		return ServiceFailedToStart, err
	}

	password, err := resolvePassword()
	if err != nil {
		return ServiceFailedToStart, err
	}

	postgresCmd, err = startPostgresProcess(port, listen, invoker)
	if err != nil {
		return ServiceFailedToStart, err
	}

	// create a RunningInfo with empty database name
	// we need this to connect to the service using 'root', required retrieve the name of the installed database
	err = createRunningInfo(postgresCmd, port, "", password, listen, invoker)
	if err != nil {
		return ServiceFailedToStart, err
	}

	databaseName, err := getDatabaseName(port)
	if err != nil {
		return ServiceFailedToStart, err
	}

	err = updateDatabaseNameInRunningInfo(databaseName)
	if err != nil {
		return ServiceFailedToStart, err
	}

	err = setServicePassword(password)
	if err != nil {
		return ServiceFailedToStart, err
	}

	// release the process - let the OS adopt it, so that we can exit
	err = postgresCmd.Process.Release()
	if err != nil {
		return ServiceFailedToStart, err
	}

	utils.LogTime("postgresCmd end")

	// ensure the foreign server exists in the database
	err = ensureSteampipeServer(databaseName)
	if err != nil {
		// there was a problem with the installation
		return ServiceFailedToStart, err
	}

	err = ensureTempTablePermissions(databaseName)
	if err != nil {
		return ServiceFailedToStart, err
	}

	// ensure the db contains command schema
	err = ensureCommandSchema(databaseName)
	if err != nil {
		return ServiceFailedToStart, err
	}

	return ServiceStarted, nil
}

// getDatabaseName connects to the service and retrieves the database name
func getDatabaseName(port int) (string, error) {
	databaseName, err := retrieveDatabaseNameFromService(port)
	if err != nil {
		return "", err
	}
	if len(databaseName) == 0 {
		return "", fmt.Errorf("could not find database to connect to")
	}
	return databaseName, nil
}

func resolvePassword() (string, error) {
	// get the password from the password file
	password, err := readPasswordFile()
	if err != nil {
		return "", err
	}

	// if a password was set through the `STEAMPIPE_DATABASE_PASSWORD` environment variable
	// or through the `--database-password` cmdline flag, then use that for this session
	// instead of the default one
	if viper.IsSet(constants.ArgServicePassword) {
		password = viper.GetString(constants.ArgServicePassword)
	}
	return password, nil
}

func startPostgresProcess(port int, listen StartListenType, invoker constants.Invoker) (*exec.Cmd, error) {
	listenAddresses := "localhost"

	if listen == ListenTypeNetwork {
		listenAddresses = "*"
	}

	if err := writePGConf(); err != nil {
		return nil, err
	}

	postgresCmd := createCmd(port, listenAddresses)

	setupLogCollection(postgresCmd)
	err := postgresCmd.Start()
	if err != nil {
		return nil, err
	}

	return postgresCmd, nil
}

func retrieveDatabaseNameFromService(port int) (string, error) {
	connection, err := createMaintenanceClient(port)
	if err != nil {
		return "", err
	}
	defer connection.Close()

	out := connection.QueryRow("select datname from pg_database where datistemplate=false AND datname <> 'postgres';")

	var databaseName string
	err = out.Scan(&databaseName)
	if err != nil {
		return "", err
	}

	return databaseName, nil
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

func updateDatabaseNameInRunningInfo(databaseName string) error {
	runningInfo, err := loadRunningInstanceInfo()
	if err != nil {
		return err
	}
	runningInfo.Database = databaseName
	return runningInfo.Save()
}

func createRunningInfo(cmd *exec.Cmd, port int, databaseName string, password string, listen StartListenType, invoker constants.Invoker) error {
	runningInfo := new(RunningDBInstanceInfo)
	runningInfo.Pid = cmd.Process.Pid
	runningInfo.Port = port
	runningInfo.User = constants.DatabaseUser
	runningInfo.Password = password
	runningInfo.Database = databaseName
	runningInfo.ListenType = listen
	runningInfo.Invoker = invoker
	runningInfo.Listen = constants.DatabaseListenAddresses

	if listen == ListenTypeNetwork {
		addrs, _ := localAddresses()
		runningInfo.Listen = append(runningInfo.Listen, addrs...)
	}
	return runningInfo.Save()
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
		"-c", fmt.Sprintf("application_name=%s", constants.AppName),
		"-c", fmt.Sprintf("cluster_name=%s", constants.AppName),

		// log directory
		"-c", fmt.Sprintf("log_directory=%s", constants.LogDir()),

		// If ssl is off  it doesnot matter what we pass in the ssl_cert_file and ssl_key_file
		// SSL will only get validated if the ssl is on
		"-c", fmt.Sprintf("ssl=%s", sslStatus()),
		"-c", fmt.Sprintf("ssl_cert_file=%s", getServerCertLocation()),
		"-c", fmt.Sprintf("ssl_key_file=%s", getServerCertKeyLocation()),

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

	// set group pgid attributes on the command to ensure the process is not shutdown when its parent terminates
	postgresCmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:    true,
		Foreground: false,
	}

	return postgresCmd
}

func setupLogCollection(cmd *exec.Cmd) {
	logChannel, stopListenFn, err := setupLogCollector(cmd)
	if err == nil {
		go traceoutServiceLogs(logChannel, stopListenFn)
	} else {
		// this is a convenience and therefore, we shouldn't error out if we
		// are not able to capture the logs.
		// instead, log to TRACE that we couldn't and continue
		log.Println("[TRACE] Warning: Could not attach to service logs")
	}
}

func traceoutServiceLogs(logChannel chan string, stopLogStreamFn func()) {
	for logLine := range logChannel {
		log.Printf("[TRACE] SERVICE: %s\n", logLine)
		if strings.Contains(logLine, "Future log output will appear in") {
			stopLogStreamFn()
			break
		}
	}
}

func setServicePassword(password string) error {
	connection, err := createLocalDbClient(&CreateDbOptions{DatabaseName: "postgres", Username: constants.DatabaseSuperUser})
	if err != nil {
		return err
	}
	defer connection.Close()
	_, err = connection.Exec(fmt.Sprintf(`alter user steampipe with password '%s'`, password))
	return err
}

func setupLogCollector(postgresCmd *exec.Cmd) (chan string, func(), error) {
	var publishChannel chan string

	stdoutPipe, err := postgresCmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	stderrPipe, err := postgresCmd.StderrPipe()
	if err != nil {
		return nil, nil, err
	}
	closeFunction := func() {
		// close the sources to make sure they don't send anymore data
		stdoutPipe.Close()
		stderrPipe.Close()

		// always close from the sender
		close(publishChannel)
	}
	stdoutScanner := bufio.NewScanner(stdoutPipe)
	stderrScanner := bufio.NewScanner(stderrPipe)

	stdoutScanner.Split(bufio.ScanLines)
	stderrScanner.Split(bufio.ScanLines)

	// create a channel with a big buffer, so that it doesn't choke
	publishChannel = make(chan string, 1000)

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

	return publishChannel, closeFunction, nil
}

// ensures that the 'steampipe' foreign server exists
//  (re)install FDW and creates server if it doesn't
func ensureSteampipeServer(databaseName string) error {
	rootClient, err := createLocalDbClient(&CreateDbOptions{DatabaseName: databaseName, Username: constants.DatabaseSuperUser})
	if err != nil {
		return err
	}
	defer rootClient.Close()
	out := rootClient.QueryRow("select srvname from pg_catalog.pg_foreign_server where srvname='steampipe'")
	var serverName string
	err = out.Scan(&serverName)
	// if there is an error, we need to reinstall the foreign server
	if err != nil {
		return installForeignServer(databaseName, rootClient)
	}
	return nil
}

// create the command schema and grant insert permission
func ensureCommandSchema(databaseName string) error {
	commandSchemaStatements := updateConnectionQuery(constants.CommandSchema, constants.CommandSchema)
	commandSchemaStatements = append(
		commandSchemaStatements,
		fmt.Sprintf("grant insert on %s.%s to steampipe_users;", constants.CommandSchema, constants.CacheCommandTable),
	)
	rootClient, err := createLocalDbClient(&CreateDbOptions{DatabaseName: databaseName, Username: constants.DatabaseSuperUser})
	if err != nil {
		return err
	}
	defer rootClient.Close()

	for _, statement := range commandSchemaStatements {
		if _, err := rootClient.Exec(statement); err != nil {
			return err
		}
	}
	return err
}

// ensures that the 'steampipe_users' role has permissions to work with temporary tables
// this is done during database installation, but we need to migrate current installations
func ensureTempTablePermissions(databaseName string) error {
	rootClient, err := createLocalDbClient(&CreateDbOptions{DatabaseName: databaseName, Username: constants.DatabaseSuperUser})
	if err != nil {
		return err
	}
	defer rootClient.Close()
	_, err = rootClient.Exec(fmt.Sprintf("grant temporary on database %s to %s", databaseName, constants.DatabaseUser))
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
		return helpers.StringSliceContains(cmdline, fmt.Sprintf("application_name=%s", constants.AppName))
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
