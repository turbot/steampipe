package db_local

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/jackc/pgx/v5"
	psutils "github.com/shirou/gopsutil/process"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/pluginmanager"
	pb "github.com/turbot/steampipe/pkg/pluginmanager_service/grpc/proto"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/sperr"
)

// StartResult is a pseudoEnum for outcomes of StartNewInstance
type StartResult struct {
	modconfig.ErrorAndWarnings
	Status             StartDbStatus
	DbState            *RunningDBInstanceInfo
	PluginManagerState *pluginmanager.PluginManagerState
}

func (r *StartResult) SetError(err error) *StartResult {
	r.Error = err
	r.Status = ServiceFailedToStart
	return r
}

// StartDbStatus is a pseudoEnum for outcomes of starting the db
type StartDbStatus int

const (
	// start from 1 to prevent confusion with int zero-value
	ServiceStarted StartDbStatus = iota + 1
	ServiceAlreadyRunning
	ServiceFailedToStart
)

// StartListenType is a pseudoEnum of network binding for postgres
type StartListenType string

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

func StartServices(ctx context.Context, port int, listen StartListenType, invoker constants.Invoker) *StartResult {
	utils.LogTime("db_local.StartServices start")
	defer utils.LogTime("db_local.StartServices end")

	res := &StartResult{}

	// if we were not successful, stop services again
	defer func() {
		if res.Status == ServiceStarted && res.Error != nil {
			StopServices(ctx, false, invoker)
			res.Status = ServiceFailedToStart
		}
	}()

	res.DbState, res.Error = GetState()
	if res.Error != nil {
		return res
	}

	if res.DbState == nil {
		res = startDB(ctx, port, listen, invoker)
		if res.Error != nil {
			return res
		}
	} else {
		res.Status = ServiceAlreadyRunning
	}

	// start plugin manager if needed
	res = ensurePluginManager(res)
	if res.Status == ServiceStarted {
		// execute post startup setup
		if err := postServiceStart(ctx, res); err != nil {
			// NOTE do not update res.Status - this will be done by defer block
			res.Error = err
		}
	}
	return res
}

func ensurePluginManager(res *StartResult) *StartResult {
	// start the plugin manager if needed
	res.PluginManagerState, res.Error = pluginmanager.LoadPluginManagerState()
	if res.Error != nil {
		res.Status = ServiceFailedToStart
		return res
	}

	if !res.PluginManagerState.Running {
		// get the location of the currently running steampipe process
		executable, err := os.Executable()
		if err != nil {
			log.Printf("[INFO] plugin manager start() - failed to get steampipe executable path: %s", err)
			return res.SetError(err)
		}
		if err := pluginmanager.StartNewInstance(executable); err != nil {
			log.Printf("[INFO] StartServices plugin manager failed to start: %s", err)
			return res.SetError(err)
		}
		// set status to service started as started plugin manager
		res.Status = ServiceStarted
	}
	return res
}

func postServiceStart(ctx context.Context, res *StartResult) error {
	conn, err := CreateLocalDbConnection(ctx, &CreateDbOptions{DatabaseName: res.DbState.Database, Username: constants.DatabaseSuperUser})
	if err != nil {
		return err
	}
	defer conn.Close(ctx)

	statushooks.SetStatus(ctx, "Dropping legacy schema")
	if err := dropLegacyInternalSchema(ctx, conn); err != nil {
		// do not fail
		// worst case scenario is that we have a couple of extra schema
		// these won't be in the search path anyway
		log.Println("[INFO] failed to drop legacy 'internal' schema", err)
	}

	// setup internal schema
	// this includes setting the state of all connections in the connection_state table to pending
	statushooks.SetStatus(ctx, "Setting up functions")
	if err := setupInternal(ctx, conn); err != nil {
		return err
	}
	// ensure connection stat etable contains entries for all connections in connection config
	// (this is to allow for the race condition between polling connection state and calling refresh connections,
	// which does not update the connection_state with added connections until it has built the ConnectionUpdates
	if err := initializeConnectionStateTable(ctx, conn); err != nil {
		return err
	}

	// create the clone_foreign_schema function
	if _, err := executeSqlAsRoot(ctx, cloneForeignSchemaSQL); err != nil {
		return sperr.WrapWithMessage(err, "failed to create clone_foreign_schema function")
	}
	// create the clone_comments function
	if _, err := executeSqlAsRoot(ctx, cloneCommentsSQL); err != nil {
		return sperr.WrapWithMessage(err, "failed to create clone_comments function")
	}

	// if there is an unprocessed db backup file, restore it now
	if err := restoreDBBackup(ctx); err != nil {
		return sperr.WrapWithMessage(err, "failed to migrate db public schema")
	}

	// call initial refresh connections
	// get plugin manager client
	pluginManager, err := pluginmanager.GetPluginManager()
	if err != nil {
		return err
	}
	// ask the plugin manager to refresh connections
	pluginManager.RefreshConnections(&pb.RefreshConnectionsRequest{})

	statushooks.SetStatus(ctx, "Service startup complete")
	return nil
}

// StartDB starts the database if not already running
func startDB(ctx context.Context, port int, listen StartListenType, invoker constants.Invoker) (res *StartResult) {
	log.Printf("[TRACE] StartDB invoker %s", invoker)
	utils.LogTime("db.StartDB start")
	defer utils.LogTime("db.StartDB end")
	var postgresCmd *exec.Cmd

	res = &StartResult{}
	defer func() {
		if r := recover(); r != nil {
			res.Error = helpers.ToError(r)
		}
		// if there was an error and we started the service, stop it again
		if res.Error != nil {
			if res.Status == ServiceStarted {
				StopServices(ctx, false, invoker)
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

	// remove the stale info file, ignoring errors - will overwrite anyway
	_ = removeRunningInstanceInfo()

	if err := utils.EnsureDirectoryPermission(getDataLocation()); err != nil {
		return res.SetError(fmt.Errorf("%s does not have the necessary permissions to start the service", getDataLocation()))
	}

	// Generate the certificate if it fails then set the ssl to off
	if err := ensureSelfSignedCertificate(); err != nil {
		error_helpers.ShowWarning("self signed certificate creation failed, connecting to the database without SSL")
	}

	if err := utils.IsPortBindable(port); err != nil {
		return res.SetError(fmt.Errorf("cannot listen on port %d", constants.Bold(port)))
	}

	if err := migrateLegacyPasswordFile(); err != nil {
		return res.SetError(err)
	}

	password, err := resolvePassword()
	if err != nil {
		return res.SetError(err)
	}

	postgresCmd, err = startPostgresProcess(ctx, port, listen, invoker)
	if err != nil {
		return res.SetError(err)
	}

	// create a RunningInfo with empty database name
	// we need this to connect to the service using 'root', required retrieve the name of the installed database
	res.DbState = newRunningDBInstanceInfo(postgresCmd, port, "", password, listen, invoker)
	err = res.DbState.Save()
	if err != nil {
		return res.SetError(err)
	}

	databaseName, err := getDatabaseName(ctx, port)
	if err != nil {
		return res.SetError(err)
	}

	res.DbState, err = updateDatabaseNameInRunningInfo(ctx, databaseName)
	if err != nil {
		return res.SetError(err)
	}

	err = setServicePassword(ctx, password)
	if err != nil {
		return res.SetError(err)
	}

	err = ensureService(ctx, databaseName)
	if err != nil {
		return res.SetError(err)
	}

	// release the process - let the OS adopt it, so that we can exit
	err = postgresCmd.Process.Release()
	if err != nil {
		return res.SetError(err)
	}

	utils.LogTime("postgresCmd end")
	res.Status = ServiceStarted
	return res
}

func ensureService(ctx context.Context, databaseName string) error {
	connection, err := CreateLocalDbConnection(ctx, &CreateDbOptions{DatabaseName: databaseName, Username: constants.DatabaseSuperUser})
	if err != nil {
		return err
	}
	defer connection.Close(ctx)

	// ensure the foreign server exists in the database
	err = ensureSteampipeServer(ctx, connection)
	if err != nil {
		return err
	}

	// ensure that the necessary extensions are installed in the database
	err = ensurePgExtensions(ctx, connection)
	if err != nil {
		// there was a problem with the installation
		return err
	}

	// ensure permissions for writing to temp tables
	err = ensureTempTablePermissions(ctx, databaseName, connection)
	if err != nil {
		return err
	}

	return nil
}

// getDatabaseName connects to the service and retrieves the database name
func getDatabaseName(ctx context.Context, port int) (string, error) {
	databaseName, err := retrieveDatabaseNameFromService(ctx, port)
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

func startPostgresProcess(ctx context.Context, port int, listen StartListenType, invoker constants.Invoker) (*exec.Cmd, error) {
	if error_helpers.IsContextCanceled(ctx) {
		return nil, ctx.Err()
	}

	listenAddresses := "localhost"

	if listen == ListenTypeNetwork {
		listenAddresses = "*"
	}

	if err := writePGConf(ctx); err != nil {
		return nil, err
	}

	postgresCmd := createCmd(ctx, port, listenAddresses)

	setupLogCollection(postgresCmd)
	err := postgresCmd.Start()
	if err != nil {
		return nil, err
	}

	return postgresCmd, nil
}

func retrieveDatabaseNameFromService(ctx context.Context, port int) (string, error) {
	connection, err := createMaintenanceClient(ctx, port)
	if err != nil {
		return "", fmt.Errorf("failed to connect to the database: %v - please try again or reset your steampipe database", err)
	}
	defer connection.Close(ctx)

	out := connection.QueryRow(ctx, "select datname from pg_database where datistemplate=false AND datname <> 'postgres';")

	var databaseName string
	err = out.Scan(&databaseName)
	if err != nil {
		return "", err
	}

	return databaseName, nil
}

func writePGConf(ctx context.Context) error {
	// Apply default settings in conf files
	err := os.WriteFile(getPostgresqlConfLocation(), []byte(constants.PostgresqlConfContent), 0600)
	if err != nil {
		return err
	}
	err = os.WriteFile(getSteampipeConfLocation(), []byte(constants.SteampipeConfContent), 0600)
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

func updateDatabaseNameInRunningInfo(ctx context.Context, databaseName string) (*RunningDBInstanceInfo, error) {
	runningInfo, err := loadRunningInstanceInfo()
	if err != nil {
		return runningInfo, err
	}
	runningInfo.Database = databaseName
	return runningInfo, runningInfo.Save()
}

func createCmd(ctx context.Context, port int, listenAddresses string) *exec.Cmd {
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
		"-c", fmt.Sprintf("log_directory=%s", filepaths.EnsureLogDir()),

		// If ssl is off  it doesnot matter what we pass in the ssl_cert_file and ssl_key_file
		// SSL will only get validated if ssl is on
		"-c", fmt.Sprintf("ssl=%s", sslStatus()),
		"-c", fmt.Sprintf("ssl_cert_file=%s", getServerCertLocation()),
		"-c", fmt.Sprintf("ssl_key_file=%s", getServerCertKeyLocation()),

		// Data Directory
		"-D", getDataLocation())

	postgresCmd.Env = append(os.Environ(), fmt.Sprintf("STEAMPIPE_INSTALL_DIR=%s", filepaths.SteampipeDir))

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

func setServicePassword(ctx context.Context, password string) error {
	connection, err := CreateLocalDbConnection(ctx, &CreateDbOptions{DatabaseName: "postgres", Username: constants.DatabaseSuperUser})
	if err != nil {
		return err
	}
	defer connection.Close(ctx)
	statements := []string{
		"LOCK TABLE pg_user IN SHARE ROW EXCLUSIVE MODE;",
		fmt.Sprintf(`ALTER USER steampipe WITH PASSWORD '%s';`, password),
	}
	_, err = ExecuteSqlInTransaction(ctx, connection, statements...)
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

// ensures that the necessary extensions are installed on the database
func ensurePgExtensions(ctx context.Context, rootClient *pgx.Conn) error {
	extensions := []string{
		"tablefunc",
		"ltree",
	}

	var errors []error
	for _, extn := range extensions {
		_, err := rootClient.Exec(ctx, fmt.Sprintf("create extension if not exists %s", db_common.PgEscapeName(extn)))
		if err != nil {
			errors = append(errors, err)
		}
	}
	return error_helpers.CombineErrors(errors...)
}

// ensures that the 'steampipe' foreign server exists
//
//	(re)install FDW and creates server if it doesn't
func ensureSteampipeServer(ctx context.Context, rootClient *pgx.Conn) error {
	res := rootClient.QueryRow(ctx, "select srvname from pg_catalog.pg_foreign_server where srvname='steampipe'")

	var serverName string
	err := res.Scan(&serverName)
	// if there is an error, we need to reinstall the foreign server
	if err != nil {
		return installForeignServer(ctx, rootClient)
	}
	return nil
}

// ensures that the 'steampipe_users' role has permissions to work with temporary tables
// this is done during database installation, but we need to migrate current installations
func ensureTempTablePermissions(ctx context.Context, databaseName string, rootClient *pgx.Conn) error {
	statements := []string{
		"lock table pg_namespace;",
		fmt.Sprintf("grant temporary on database %s to %s", databaseName, constants.DatabaseUser),
	}
	if _, err := ExecuteSqlInTransaction(ctx, rootClient, statements...); err != nil {
		return err
	}
	return nil
}

// kill all postgres processes that were started as part of steampipe (if any)
func killInstanceIfAny(ctx context.Context) bool {
	processes, err := FindAllSteampipePostgresInstances(ctx)
	if err != nil {
		return false
	}
	wg := sync.WaitGroup{}
	for _, process := range processes {
		wg.Add(1)
		go func(p *psutils.Process) {
			doThreeStepPostgresExit(ctx, p)
			wg.Done()
		}(process)
	}
	wg.Wait()
	return len(processes) > 0
}

func FindAllSteampipePostgresInstances(ctx context.Context) ([]*psutils.Process, error) {
	var instances []*psutils.Process
	allProcesses, err := psutils.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range allProcesses {
		cmdLine, err := p.CmdlineSliceWithContext(ctx)
		if err != nil {
			return nil, err
		}
		if isSteampipePostgresProcess(ctx, cmdLine) {
			instances = append(instances, p)
		}
	}
	return instances, nil
}

func isSteampipePostgresProcess(ctx context.Context, cmdline []string) bool {
	if len(cmdline) < 1 {
		return false
	}
	if strings.Contains(cmdline[0], "postgres") {
		// this is a postgres process - but is it a steampipe service?
		return helpers.StringSliceContains(cmdline, fmt.Sprintf("application_name=%s", constants.AppName))
	}
	return false
}
