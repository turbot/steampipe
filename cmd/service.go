package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardserver"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/display"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/pluginmanager"
	pb "github.com/turbot/steampipe/pkg/pluginmanager_service/grpc/proto"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/utils"
)

func serviceCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "service [command]",
		Args:  cobra.NoArgs,
		Short: "Steampipe service management",
		Long: `Steampipe service management.

Run Steampipe as a local service, exposing it as a database endpoint for
connection from any Postgres compatible database client.`,
	}

	cmd.AddCommand(serviceStartCmd())
	cmd.AddCommand(serviceStatusCmd())
	cmd.AddCommand(serviceStopCmd())
	cmd.AddCommand(serviceRestartCmd())
	cmd.Flags().BoolP(constants.ArgHelp, "h", false, "Help for service")
	return cmd
}

// handler for service start
func serviceStartCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "start",
		Args:  cobra.NoArgs,
		Run:   runServiceStartCmd,
		Short: "Start Steampipe in service mode",
		Long: `Start the Steampipe service.

Run Steampipe as a local service, exposing it as a database endpoint for
connection from any Postgres compatible database client.`,
	}

	cmdconfig.
		OnCmd(cmd).
		AddModLocationFlag().
		AddBoolFlag(constants.ArgHelp, false, "Help for service start", cmdconfig.FlagOptions.WithShortHand("h")).
		AddIntFlag(constants.ArgDatabasePort, constants.DatabaseDefaultPort, "Database service port").
		AddStringFlag(constants.ArgDatabaseListenAddresses, string(db_local.ListenTypeNetwork), "Accept connections from: `local` (an alias for `localhost` only), `network` (an alias for `*`), or a comma separated list of hosts and/or IP addresses").
		AddStringFlag(constants.ArgServicePassword, "", "Set the database password for this session").
		// default is false and hides the database user password from service start prompt
		AddBoolFlag(constants.ArgServiceShowPassword, false, "View database password for connecting from another machine").
		// dashboard server
		AddBoolFlag(constants.ArgDashboard, false, "Run the dashboard webserver with the service").
		AddStringFlag(constants.ArgDashboardListen, string(dashboardserver.ListenTypeNetwork), "Accept connections from: local (localhost only) or network (open) (dashboard)").
		AddIntFlag(constants.ArgDashboardPort, constants.DashboardServerDefaultPort, "Report server port").
		// foreground enables the service to run in the foreground - till exit
		AddBoolFlag(constants.ArgForeground, false, "Run the service in the foreground").

		// flags relevant only if the --dashboard arg is used:
		AddStringSliceFlag(constants.ArgVarFile, nil, "Specify an .spvar file containing variable values (only applies if '--dashboard' flag is also set)").
		// NOTE: use StringArrayFlag for ArgVariable, not StringSliceFlag
		// Cobra will interpret values passed to a StringSliceFlag as CSV,
		// where args passed to StringArrayFlag are not parsed and used raw
		AddStringArrayFlag(constants.ArgVariable, nil, "Specify the value of a variable (only applies if '--dashboard' flag is also set)").

		// hidden flags for internal use
		AddStringFlag(constants.ArgInvoker, string(constants.InvokerService), "Invoked by \"service\" or \"query\"", cmdconfig.FlagOptions.Hidden())

	return cmd
}

// serviceStatusCmd :: handler for service status
func serviceStatusCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "status",
		Args:  cobra.NoArgs,
		Run:   runServiceStatusCmd,
		Short: "Status of the Steampipe service",
		Long: `Status of the Steampipe service.

Report current status of the Steampipe database service.`,
	}

	cmdconfig.OnCmd(cmd).
		AddBoolFlag(constants.ArgHelp, false, "Help for service status", cmdconfig.FlagOptions.WithShortHand("h")).
		// default is false and hides the database user password from service start prompt
		AddBoolFlag(constants.ArgServiceShowPassword, false, "View database password for connecting from another machine").
		AddBoolFlag(constants.ArgAll, false, "Bypasses the INSTALL_DIR and reports status of all running steampipe services")

	return cmd
}

// handler for service stop
func serviceStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Args:  cobra.NoArgs,
		Run:   runServiceStopCmd,
		Short: "Stop Steampipe service",
		Long:  `Stop the Steampipe service.`,
	}

	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag(constants.ArgHelp, false, "Help for service stop", cmdconfig.FlagOptions.WithShortHand("h")).
		AddBoolFlag(constants.ArgForce, false, "Forces all services to shutdown, releasing all open connections and ports")

	return cmd
}

// restarts the database service
func serviceRestartCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "restart",
		Args:  cobra.NoArgs,
		Run:   runServiceRestartCmd,
		Short: "Restart Steampipe service",
		Long:  `Restart the Steampipe service.`,
	}

	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag(constants.ArgHelp, false, "Help for service restart", cmdconfig.FlagOptions.WithShortHand("h")).
		AddBoolFlag(constants.ArgForce, false, "Forces the service to restart, releasing all open connections and ports")

	return cmd
}

func runServiceStartCmd(cmd *cobra.Command, _ []string) {
	ctx := cmd.Context()
	utils.LogTime("runServiceStartCmd start")
	defer func() {
		utils.LogTime("runServiceStartCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			if exitCode == constants.ExitCodeSuccessful {
				// there was an error and the exitcode
				// was not set to a non-zero value.
				// set it
				exitCode = constants.ExitCodeUnknownErrorPanic
			}
		}
	}()

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()

	listenAddresses := db_local.StartListenType(viper.GetString(constants.ArgDatabaseListenAddresses)).ToListenAddresses()

	port := viper.GetInt(constants.ArgDatabasePort)
	if port < 1 || port > 65535 {
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		panic("Invalid port - must be within range (1:65535)")
	}

	invoker := constants.Invoker(cmdconfig.Viper().GetString(constants.ArgInvoker))
	if invoker.IsValid() != nil {
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		error_helpers.FailOnError(invoker.IsValid())
	}

	startResult, dashboardState, dbServiceStarted := startService(ctx, listenAddresses, port, invoker)
	alreadyRunning := !dbServiceStarted

	printStatus(ctx, startResult.DbState, startResult.PluginManagerState, dashboardState, alreadyRunning)

	if viper.GetBool(constants.ArgForeground) {
		runServiceInForeground(ctx)
	}
}

func startService(ctx context.Context, listenAddresses []string, port int, invoker constants.Invoker) (_ *db_local.StartResult, _ *dashboardserver.DashboardServiceState, dbServiceStarted bool) {
	statushooks.Show(ctx)
	defer statushooks.Done(ctx)
	log.Printf("[TRACE] startService - listenAddresses=%q", listenAddresses)

	err := db_local.EnsureDBInstalled(ctx)
	if err != nil {
		exitCode = constants.ExitCodeServiceStartupFailure
		error_helpers.FailOnError(err)
	}

	// start db, refreshing connections
	startResult := startServiceAndRefreshConnections(ctx, listenAddresses, port, invoker)
	if startResult.Status == db_local.ServiceFailedToStart {
		error_helpers.ShowError(ctx, sperr.New("steampipe service failed to start"))
		exitCode = constants.ExitCodeServiceStartupFailure
		return
	}

	// if the service is already running, then service start should make the service persistent
	if startResult.Status == db_local.ServiceAlreadyRunning {
		// check that we have the same port and listen parameters
		if port != startResult.DbState.Port {
			exitCode = constants.ExitCodeInsufficientOrWrongInputs
			error_helpers.FailOnError(sperr.New("service is already running on port %d - cannot change port while it's running", startResult.DbState.Port))
		}
		if !startResult.DbState.MatchWithGivenListenAddresses(listenAddresses) {
			exitCode = constants.ExitCodeInsufficientOrWrongInputs
			// this messaging assumes that the resolved addresses from the given addresses have not changed while the service is running
			// although this is an edge case, ideally, we should check for the resolved addresses and give the relevant message
			error_helpers.FailOnError(sperr.New("service is already running and listening on %s - cannot change listen address while it's running", strings.Join(startResult.DbState.ResolvedListenAddresses, ", ")))
		}

		// convert to being invoked by service
		startResult.DbState.Invoker = constants.InvokerService
		err = startResult.DbState.Save()
		if err != nil {
			exitCode = constants.ExitCodeFileSystemAccessFailure
			error_helpers.FailOnErrorWithMessage(err, "service was already running, but could not make it persistent")
		}
	}

	dbServiceStarted = startResult.Status == db_local.ServiceStarted

	var dashboardState *dashboardserver.DashboardServiceState
	if viper.GetBool(constants.ArgDashboard) {
		dashboardState, err = dashboardserver.GetDashboardServiceState()
		if err != nil {
			tryToStopServices(ctx)
			exitCode = constants.ExitCodeServiceStartupFailure
			error_helpers.FailOnError(err)
		}
		if dashboardState == nil {
			dashboardState, err = startDashboardServer(ctx)
			if err != nil {
				tryToStopServices(ctx)
				exitCode = constants.ExitCodeServiceStartupFailure
				error_helpers.FailOnError(err)
			}
			dbServiceStarted = true
		}
	}
	return startResult, dashboardState, dbServiceStarted
}

func startServiceAndRefreshConnections(ctx context.Context, listenAddresses []string, port int, invoker constants.Invoker) *db_local.StartResult {
	startResult := db_local.StartServices(ctx, listenAddresses, port, invoker)
	if startResult.Error != nil {
		exitCode = constants.ExitCodeServiceStartupFailure
		error_helpers.FailOnError(startResult.Error)
	}

	if startResult.Status == db_local.ServiceStarted {
		// ask the plugin manager to refresh connections
		// this is executed asyncronously by the plugin manager
		// we ignore this error, since RefreshConnections is async and all errors will flow through
		// the notification system
		// we do not expect any I/O errors on this since the PluginManager is running in the same box
		_, _ = startResult.PluginManager.RefreshConnections(&pb.RefreshConnectionsRequest{})
	}
	return startResult
}

func tryToStopServices(ctx context.Context) {
	// stop db service
	if _, err := db_local.StopServices(ctx, false, constants.InvokerService); err != nil {
		error_helpers.ShowError(ctx, err)
	}
	// stop the dashboard service
	if err := dashboardserver.StopDashboardService(ctx); err != nil {
		error_helpers.ShowError(ctx, err)
	}
}

func startDashboardServer(ctx context.Context) (*dashboardserver.DashboardServiceState, error) {
	var dashboardState *dashboardserver.DashboardServiceState
	var err error

	serverPort := dashboardserver.ListenPort(viper.GetInt(constants.ArgDashboardPort))
	serverListen := dashboardserver.ListenType(viper.GetString(constants.ArgDashboardListen))

	dashboardState, err = dashboardserver.GetDashboardServiceState()
	if err != nil {
		return nil, err
	}

	if dashboardState == nil {
		// try stopping the previous service
		// StopDashboardService does nothing if the service is not running
		err = dashboardserver.StopDashboardService(ctx)
		if err != nil {
			return nil, err
		}
		// start dashboard service
		err = dashboardserver.RunForService(ctx, serverListen, serverPort)
		if err != nil {
			return nil, err
		}
		// get the updated state
		dashboardState, err = dashboardserver.GetDashboardServiceState()
		if err != nil {
			error_helpers.ShowWarning(fmt.Sprintf("Started Dashboard server, but could not retrieve state: %v", err))
		}
	}

	return dashboardState, err
}

func runServiceInForeground(ctx context.Context) {
	fmt.Println("Hit Ctrl+C to stop the service")

	sigIntChannel := make(chan os.Signal, 1)
	signal.Notify(sigIntChannel, os.Interrupt)

	checkTimer := time.NewTicker(100 * time.Millisecond)
	defer checkTimer.Stop()

	var lastCtrlC time.Time

	for {
		select {
		case <-checkTimer.C:
			// get the current status
			newInfo, err := db_local.GetState()
			if err != nil {
				continue
			}
			if newInfo == nil {
				fmt.Println("Steampipe service stopped.")
				return
			}
		case <-sigIntChannel:
			fmt.Print("\r")
			dashboardserver.StopDashboardService(ctx)
			// if we have received this signal, then the user probably wants to shut down
			// everything. Shutdowns MUST NOT happen in cancellable contexts
			connectedClients, err := db_local.GetClientCount(context.Background())
			if err != nil {
				// report the error in the off chance that there's one
				error_helpers.ShowError(ctx, err)
				return
			}

			// we know there will be at least 1 client (connectionWatcher)
			if connectedClients.TotalClients > 1 {
				if lastCtrlC.IsZero() || time.Since(lastCtrlC) > 30*time.Second {
					lastCtrlC = time.Now()
					fmt.Println(buildForegroundClientsConnectedMsg())
					continue
				}
			}
			fmt.Println("Stopping Steampipe service.")
			if _, err := db_local.StopServices(ctx, false, constants.InvokerService); err != nil {
				error_helpers.ShowError(ctx, err)
			} else {
				fmt.Println("Steampipe service stopped.")
			}
			return
		}
	}
}

func runServiceRestartCmd(cmd *cobra.Command, _ []string) {
	ctx := cmd.Context()
	utils.LogTime("runServiceRestartCmd start")
	defer func() {
		utils.LogTime("runServiceRestartCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			if exitCode == constants.ExitCodeSuccessful {
				// there was an error and the exitcode
				// was not set to a non-zero value.
				// set it
				exitCode = constants.ExitCodeUnknownErrorPanic
			}
		}
	}()

	dbStartResult, currentDashboardState := restartService(ctx)

	if dbStartResult != nil {
		printStatus(ctx, dbStartResult.DbState, dbStartResult.PluginManagerState, currentDashboardState, false)
	}
}

func restartService(ctx context.Context) (_ *db_local.StartResult, _ *dashboardserver.DashboardServiceState) {
	statushooks.Show(ctx)
	defer statushooks.Done(ctx)

	// get current db statue
	currentDbState, err := db_local.GetState()
	error_helpers.FailOnError(err)
	if currentDbState == nil {
		fmt.Println("Steampipe service is not running.")
		return
	}

	// along with the current dashboard state - maybe nil
	currentDashboardState, err := dashboardserver.GetDashboardServiceState()
	error_helpers.FailOnError(err)

	// stop db
	stopStatus, err := db_local.StopServices(ctx, viper.GetBool(constants.ArgForce), constants.InvokerService)
	if err != nil {
		exitCode = constants.ExitCodeServiceStopFailure
		error_helpers.FailOnErrorWithMessage(err, "could not stop current instance")
	}

	if stopStatus != db_local.ServiceStopped {
		fmt.Println(`
Service stop failed.

Try using:
	steampipe service restart --force

to force a restart.
		`)
		return
	}

	// stop the running dashboard server
	err = dashboardserver.StopDashboardService(ctx)
	if err != nil {
		exitCode = constants.ExitCodeServiceStopFailure
		error_helpers.FailOnErrorWithMessage(err, "could not stop dashboard service")
	}

	// the DB must be installed and therefore is a noop,
	// and EnsureDBInstalled also checks and installs the latest FDW
	err = db_local.EnsureDBInstalled(ctx)
	if err != nil {
		exitCode = constants.ExitCodeServiceStartupFailure
		error_helpers.FailOnError(err)
	}

	// set the password in 'viper' so that it can be used by 'service start'
	viper.Set(constants.ArgServicePassword, currentDbState.Password)

	// start db
	dbStartResult := startServiceAndRefreshConnections(ctx, currentDbState.ResolvedListenAddresses, currentDbState.Port, currentDbState.Invoker)
	if dbStartResult.Status == db_local.ServiceFailedToStart {
		exitCode = constants.ExitCodeServiceStartupFailure
		fmt.Println("Steampipe service was stopped, but failed to restart.")
		return
	}

	// if the dashboard was running, start it
	if currentDashboardState != nil {
		err = dashboardserver.RunForService(ctx, dashboardserver.ListenType(currentDashboardState.ListenType), dashboardserver.ListenPort(currentDashboardState.Port))
		error_helpers.FailOnError(err)

		// reload the state
		currentDashboardState, err = dashboardserver.GetDashboardServiceState()
		error_helpers.FailOnError(err)
	}

	return dbStartResult, currentDashboardState
}

func runServiceStatusCmd(cmd *cobra.Command, _ []string) {
	ctx := cmd.Context()
	utils.LogTime("runServiceStatusCmd status")
	defer func() {
		utils.LogTime("runServiceStatusCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
		}
	}()

	if !db_local.IsDBInstalled() || !db_local.IsFDWInstalled() {
		fmt.Println("Steampipe service is not installed.")
		return
	}

	if viper.GetBool(constants.ArgAll) {
		showAllStatus(ctx)
	} else {
		dbState, dbStateErr := db_local.GetState()
		pmState, pmStateErr := pluginmanager.LoadState()
		dashboardState, dashboardStateErr := dashboardserver.GetDashboardServiceState()

		if dbStateErr != nil || pmStateErr != nil {
			error_helpers.ShowError(ctx, composeStateError(dbStateErr, pmStateErr, dashboardStateErr))
			return
		}
		printStatus(ctx, dbState, pmState, dashboardState, false)
	}
}

func composeStateError(dbStateErr error, pmStateErr error, dashboardStateErr error) error {
	msg := "could not get Steampipe service status:"

	if dbStateErr != nil {
		msg = fmt.Sprintf(`%s
	failed to get db state: %s`, msg, dbStateErr.Error())
	}
	if pmStateErr != nil {
		msg = fmt.Sprintf(`%s
	failed to get plugin manager state: %s`, msg, pmStateErr.Error())
	}
	if dashboardStateErr != nil {
		msg = fmt.Sprintf(`%s
	failed to get dashboard server state: %s`, msg, pmStateErr.Error())
	}

	return errors.New(msg)
}

func runServiceStopCmd(cmd *cobra.Command, _ []string) {
	ctx := cmd.Context()
	utils.LogTime("runServiceStopCmd stop")

	var status db_local.StopStatus
	var dbStopError error
	var dbState *db_local.RunningDBInstanceInfo

	defer func() {
		utils.LogTime("runServiceStopCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
			if exitCode == constants.ExitCodeSuccessful {
				// there was an error and the exitcode
				// was not set to a non-zero value.
				// set it
				exitCode = constants.ExitCodeUnknownErrorPanic
			}
		}
	}()

	force := cmdconfig.Viper().GetBool(constants.ArgForce)
	if force {
		dashboardStopError := dashboardserver.StopDashboardService(ctx)
		status, dbStopError = db_local.StopServices(ctx, force, constants.InvokerService)
		dbStopError = error_helpers.CombineErrors(dbStopError, dashboardStopError)
		if dbStopError != nil {
			exitCode = constants.ExitCodeServiceStopFailure
			error_helpers.FailOnError(dbStopError)
		}
	} else {
		dbState, dbStopError = db_local.GetState()
		if dbStopError != nil {
			exitCode = constants.ExitCodeServiceStopFailure
			error_helpers.FailOnErrorWithMessage(dbStopError, "could not stop Steampipe service")
		}

		dashboardState, err := dashboardserver.GetDashboardServiceState()
		if err != nil {
			exitCode = constants.ExitCodeServiceStopFailure
			error_helpers.FailOnErrorWithMessage(err, "could not stop Steampipe service")
		}

		if dbState == nil {
			fmt.Println("Steampipe service is not running.")
			return
		}
		if dbState.Invoker != constants.InvokerService {
			printRunningImplicit(dbState.Invoker)
			return
		}

		if dashboardState != nil {
			err = dashboardserver.StopDashboardService(ctx)
			if err != nil {
				exitCode = constants.ExitCodeServiceStopFailure
				error_helpers.FailOnErrorWithMessage(err, "could not stop dashboard server")
			}
		}

		// check if there are any connected clients to the service
		connectedClients, err := db_local.GetClientCount(ctx)
		if err != nil {
			exitCode = constants.ExitCodeServiceStopFailure
			error_helpers.FailOnErrorWithMessage(err, "service stop failed")
		}

		// if there are any clients connected (apart from plugin manager clients), do not exit
		if connectedClients.TotalClients-connectedClients.PluginManagerClients > 0 {
			printClientsConnected()
			return
		}

		status, err = db_local.StopServices(ctx, false, constants.InvokerService)
		if err != nil {
			exitCode = constants.ExitCodeServiceStopFailure
			error_helpers.FailOnErrorWithMessage(err, "service stop failed")
		}
	}

	switch status {
	case db_local.ServiceStopped:
		fmt.Println("Steampipe database service stopped.")
	case db_local.ServiceNotRunning:
		fmt.Println("Steampipe service is not running.")
	case db_local.ServiceStopFailed:
		fmt.Println("Could not stop Steampipe service.")
	case db_local.ServiceStopTimedOut:
		fmt.Println(`
Service stop operation timed-out.

This is probably because other clients are connected to the database service.

Disconnect all clients, or use
	steampipe service stop --force

to force a shutdown.
		`)

	}
}

func showAllStatus(ctx context.Context) {
	var processes []*psutils.Process
	var err error

	statushooks.SetStatus(ctx, "Getting details")
	processes, err = db_local.FindAllSteampipePostgresInstances(ctx)
	statushooks.Done(ctx)

	error_helpers.FailOnError(err)

	if len(processes) == 0 {
		fmt.Println("There are no steampipe services running.")
		return
	}
	headers := []string{"PID", "Install Directory", "Port", "Listen"}
	rows := [][]string{}

	for _, process := range processes {
		pid, installDir, port, listen := getServiceProcessDetails(process)
		rows = append(rows, []string{pid, installDir, port, string(listen)})
	}

	display.ShowWrappedTable(headers, rows, &display.ShowWrappedTableOptions{AutoMerge: false})
}

func getServiceProcessDetails(process *psutils.Process) (string, string, string, db_local.StartListenType) {
	cmdLine, _ := process.CmdlineSlice()
	installDir := strings.TrimSuffix(cmdLine[0], filepaths.ServiceExecutableRelativeLocation())
	var port string
	var listenType db_local.StartListenType

	for idx, param := range cmdLine {
		if param == "-p" {
			port = cmdLine[idx+1]
		}
		if strings.HasPrefix(param, "listen_addresses") {
			if strings.Contains(param, "localhost") {
				listenType = db_local.ListenTypeLocal
			} else {
				listenType = db_local.ListenTypeNetwork
			}
		}
	}

	return fmt.Sprintf("%d", process.Pid), installDir, port, listenType
}

func printStatus(ctx context.Context, dbState *db_local.RunningDBInstanceInfo, pmState *pluginmanager.State, dashboardState *dashboardserver.DashboardServiceState, alreadyRunning bool) {
	if dbState == nil && !pmState.Running {
		fmt.Println("Service is not running")
		return
	}

	var statusMessage string

	prefix := `Steampipe service is running:
`
	if alreadyRunning {
		prefix = `Steampipe service is already running:
`
	}
	suffix := `
Managing the Steampipe service:

  # Get status of the service
  steampipe service status

  # View database password for connecting from another machine
  steampipe service status --show-password

  # Restart the service
  steampipe service restart

  # Stop the service
  steampipe service stop
`

	var connectionStr string
	var password string
	if viper.GetBool(constants.ArgServiceShowPassword) {
		connectionStr = fmt.Sprintf(
			"postgres://%v:%v@%v:%v/%v",
			dbState.User,
			dbState.Password,
			utils.GetFirstListenAddress(dbState.ResolvedListenAddresses),
			dbState.Port,
			dbState.Database,
		)
		password = dbState.Password
	} else {
		connectionStr = fmt.Sprintf(
			"postgres://%v@%v:%v/%v",
			dbState.User,
			utils.GetFirstListenAddress(dbState.ResolvedListenAddresses),
			dbState.Port,
			dbState.Database,
		)
		password = "********* [use --show-password to reveal]"
	}

	postgresFmt := `
Database:

  Host(s):            %v
  Port:               %v
  Database:           %v
  User:               %v
  Password:           %v
  Connection string:  %v
`
	postgresMsg := fmt.Sprintf(
		postgresFmt,
		strings.Join(dbState.ResolvedListenAddresses, ", "),
		dbState.Port,
		dbState.Database,
		dbState.User,
		password,
		connectionStr,
	)

	dashboardMsg := ""

	if dashboardState != nil {
		browserUrl := fmt.Sprintf("http://%s:%d/", dashboardState.Listen[0], dashboardState.Port)
		dashboardMsg = fmt.Sprintf(`
Dashboard:

  Host(s):  %v
  Port:     %v
  URL:      %v
`, strings.Join(dashboardState.Listen, ", "), dashboardState.Port, browserUrl)
	}

	if dbState.Invoker == constants.InvokerService {
		statusMessage = fmt.Sprintf(
			"%s%s%s%s",
			prefix,
			postgresMsg,
			dashboardMsg,
			suffix,
		)
	} else {
		msg := `
Steampipe service was started for an active %s session. The service will exit when all active sessions exit.

To keep the service running after the %s session completes, use %s.
`

		statusMessage = fmt.Sprintf(
			msg,
			fmt.Sprintf("steampipe %s", dbState.Invoker),
			dbState.Invoker,
			constants.Bold("steampipe service start"),
		)
	}

	fmt.Println(statusMessage)

	if dbState != nil && pmState == nil {
		// the service is running, but the plugin_manager is not running and there's no state file
		// meaning that it cannot be restarted by the FDW
		// it's an ERROR
		error_helpers.ShowError(ctx, sperr.New(`
Service is running, but the Plugin Manager cannot be recovered.
Please use %s to recover the service
`,
			constants.Bold("steampipe service restart"),
		))
	}
}

func printRunningImplicit(invoker constants.Invoker) {
	fmt.Printf(`
Steampipe service is running exclusively for an active %s session.

To force stop the service, use %s

`,
		fmt.Sprintf("steampipe %s", invoker),
		constants.Bold("steampipe service stop --force"),
	)
}

func printClientsConnected() {
	fmt.Printf(
		`
Cannot stop service since there are clients connected to the service.

To force stop the service, use %s

`,
		constants.Bold("steampipe service stop --force"),
	)
}

func buildForegroundClientsConnectedMsg() string {
	return `
Not shutting down service as there as clients connected.

To force shutdown, press Ctrl+C again.
	`
}
