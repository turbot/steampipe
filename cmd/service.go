package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardserver"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/display"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pluginmanager"
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
		AddBoolFlag(constants.ArgHelp, false, "Help for service start", cmdconfig.FlagOptions.WithShortHand("h")).
		// for now default port to -1 so we fall back to the default of the deprecated arg
		AddIntFlag(constants.ArgDatabasePort, constants.DatabaseDefaultPort, "Database service port").
		// for now default listen address to empty so we fall back to the default of the deprecated arg
		AddStringFlag(constants.ArgListenAddress, string(db_local.ListenTypeNetwork), "Accept connections from: local (localhost only) or network (open) (postgres)").
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

func runServiceStartCmd(cmd *cobra.Command, args []string) {
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

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, os.Kill)
	defer cancel()

	port := viper.GetInt(constants.ArgDatabasePort)
	if port < 1 || port > 65535 {
		panic("Invalid port - must be within range (1:65535)")
	}

	serviceListen := db_local.StartListenType(viper.GetString(constants.ArgListenAddress))
	error_helpers.FailOnError(serviceListen.IsValid())

	invoker := constants.Invoker(cmdconfig.Viper().GetString(constants.ArgInvoker))
	error_helpers.FailOnError(invoker.IsValid())

	err := db_local.EnsureDBInstalled(ctx)
	error_helpers.FailOnError(err)

	// start db, refreshing connections
	startResult := db_local.StartServices(ctx, port, serviceListen, invoker)
	error_helpers.FailOnError(startResult.Error)

	if startResult.Status == db_local.ServiceFailedToStart {
		error_helpers.ShowError(ctx, fmt.Errorf("steampipe service failed to start"))
		return
	}

	if startResult.Status == db_local.ServiceAlreadyRunning {

		// check that we have the same port and listen parameters
		if port != startResult.DbState.Port {
			error_helpers.FailOnError(fmt.Errorf("service is already running on port %d - cannot change port while it's running", startResult.DbState.Port))
		}
		if serviceListen != startResult.DbState.ListenType {
			error_helpers.FailOnError(fmt.Errorf("service is already running and listening on %s - cannot change listen type while it's running", startResult.DbState.ListenType))
		}

		// convert to being invoked by service
		startResult.DbState.Invoker = constants.InvokerService
		err = startResult.DbState.Save()
		if err != nil {
			error_helpers.FailOnErrorWithMessage(err, "service was already running, but could not make it persistent")
		}
	}

	// if the service was started
	if startResult.Status == db_local.ServiceStarted {
		// do
		err = db_local.RefreshConnectionAndSearchPaths(ctx, invoker)
		if err != nil {
			_, err1 := db_local.StopServices(ctx, false, constants.InvokerService)
			if err1 != nil {
				error_helpers.ShowError(ctx, err1)
			}
			error_helpers.FailOnError(err)
		}
	}

	servicesStarted := startResult.Status == db_local.ServiceStarted

	var dashboardState *dashboardserver.DashboardServiceState
	if viper.GetBool(constants.ArgDashboard) {
		dashboardState, err = dashboardserver.GetDashboardServiceState()
		if err != nil {
			tryToStopServices(ctx)
			error_helpers.ShowError(ctx, err)
			return
		}
		if dashboardState == nil {
			dashboardState, err = startDashboardServer(ctx)
			if err != nil {
				error_helpers.ShowError(ctx, err)
				tryToStopServices(ctx)
				return
			}
			servicesStarted = true
		}
	}

	printStatus(ctx, startResult.DbState, startResult.PluginManagerState, dashboardState, !servicesStarted)

	if viper.GetBool(constants.ArgForeground) {
		runServiceInForeground(ctx, invoker)
	}
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

func runServiceInForeground(ctx context.Context, invoker constants.Invoker) {
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
			count, err := db_local.GetCountOfThirdPartyClients(context.Background())
			if err != nil {
				// report the error in the off chance that there's one
				error_helpers.ShowError(ctx, err)
				return
			}

			// we know there will be at least 1 client (connectionWatcher)
			if count > 1 {
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

func runServiceRestartCmd(cmd *cobra.Command, args []string) {
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
	error_helpers.FailOnErrorWithMessage(err, "could not stop current instance")
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
	error_helpers.FailOnErrorWithMessage(err, "could not stop dashboard service")

	// set the password in 'viper' so that it can be used by 'service start'
	viper.Set(constants.ArgServicePassword, currentDbState.Password)

	// start db
	dbStartResult := db_local.StartServices(cmd.Context(), currentDbState.Port, currentDbState.ListenType, currentDbState.Invoker)
	error_helpers.FailOnError(dbStartResult.Error)
	if dbStartResult.Status == db_local.ServiceFailedToStart {
		fmt.Println("Steampipe service was stopped, but failed to restart.")
		return
	}

	// refresh connections
	err = db_local.RefreshConnectionAndSearchPaths(cmd.Context(), constants.InvokerService)
	error_helpers.FailOnError(err)

	// if the dashboard was running, start it
	if currentDashboardState != nil {
		err = dashboardserver.RunForService(ctx, dashboardserver.ListenType(currentDashboardState.ListenType), dashboardserver.ListenPort(currentDashboardState.Port))
		error_helpers.FailOnError(err)

		// reload the state
		currentDashboardState, err = dashboardserver.GetDashboardServiceState()
		error_helpers.FailOnError(err)
	}

	printStatus(ctx, dbStartResult.DbState, dbStartResult.PluginManagerState, currentDashboardState, false)
}

func runServiceStatusCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("runServiceStatusCmd status")
	defer func() {
		utils.LogTime("runServiceStatusCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
		}
	}()

	if !db_local.IsInstalled() {
		fmt.Println("Steampipe service is not installed.")
		return
	}
	if viper.GetBool(constants.ArgAll) {
		showAllStatus(cmd.Context())
	} else {
		dbState, dbStateErr := db_local.GetState()
		pmState, pmStateErr := pluginmanager.LoadPluginManagerState()
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

	return errors.New(msg)
}

func runServiceStopCmd(cmd *cobra.Command, args []string) {
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
		error_helpers.FailOnError(dbStopError)
	} else {
		dbState, dbStopError = db_local.GetState()
		error_helpers.FailOnErrorWithMessage(dbStopError, "could not stop Steampipe service")

		dashboardState, err := dashboardserver.GetDashboardServiceState()
		error_helpers.FailOnErrorWithMessage(err, "could not stop Steampipe service")

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
			error_helpers.FailOnErrorWithMessage(err, "could not stop dashboard server")
		}

		var connectedClientCount int
		// check if there are any connected clients to the service
		connectedClientCount, err = db_local.GetCountOfThirdPartyClients(cmd.Context())
		error_helpers.FailOnErrorWithMessage(err, "service stop failed")

		if connectedClientCount > 0 {
			printClientsConnected()
			return
		}

		status, err = db_local.StopServices(ctx, false, constants.InvokerService)
		error_helpers.FailOnErrorWithMessage(err, "service stop failed")
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
	installDir := strings.TrimSuffix(cmdLine[0], db_local.ServiceExecutableRelativeLocation())
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

func printStatus(ctx context.Context, dbState *db_local.RunningDBInstanceInfo, pmState *pluginmanager.PluginManagerState, dashboardState *dashboardserver.DashboardServiceState, alreadyRunning bool) {
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
			dbState.Listen[0],
			dbState.Port,
			dbState.Database,
		)
		password = dbState.Password
	} else {
		connectionStr = fmt.Sprintf(
			"postgres://%v@%v:%v/%v",
			dbState.User,
			dbState.Listen[0],
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
		strings.Join(dbState.Listen, ", "),
		dbState.Port,
		dbState.Database,
		dbState.User,
		password,
		connectionStr,
	)

	dashboardMsg := ""

	if dashboardState != nil {
		browserUrl := fmt.Sprintf("http://localhost:%d/", dashboardState.Port)
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
		error_helpers.ShowError(ctx, fmt.Errorf(`
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
