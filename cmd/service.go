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
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_local"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/plugin_manager"
	"github.com/turbot/steampipe/utils"
)

// serviceCmd :: Service management commands
func serviceCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "service [command]",
		Args:  cobra.NoArgs,
		Short: "Steampipe service management",
		// TODO(nw) expand long description
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

// serviceStartCmd :: handler for service start
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
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for service start").
		// for now default port to -1 so we fall back to the default of the deprecated arg
		AddIntFlag(constants.ArgPort, "", constants.DatabaseDefaultPort, "Database service port.").
		// for now default listen address to empty so we fall back to the default of the deprecated arg
		AddStringFlag(constants.ArgListenAddress, "", string(db_local.ListenTypeNetwork), "Accept connections from: local (localhost only) or network (open)").
		AddStringFlag(constants.ArgServicePassword, "", "", "Set the database password for this session").
		// foreground enables the service to run in the foreground - till exit
		AddBoolFlag(constants.ArgForeground, "", false, "Run the service in the foreground").
		// Hidden flags for internal use
		AddStringFlag(constants.ArgInvoker, "", string(constants.InvokerService), "Invoked by \"service\" or \"query\"", cmdconfig.FlagOptions.Hidden())

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
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for service status").
		AddBoolFlag(constants.ArgAll, "", false, "Bypasses the INSTALL_DIR and reports status of all running steampipe services")

	return cmd
}

// serviceStopCmd :: handler for service stop
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
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for service stop").
		AddBoolFlag(constants.ArgForce, "", false, "Forces all services to shutdown, releasing all open connections and ports")

	return cmd
}

// serviceRestartCmd :: restarts the database service
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
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for service restart").
		AddBoolFlag(constants.ArgForce, "", false, "Forces the service to restart, releasing all open connections and ports")

	return cmd
}

func runServiceStartCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("runServiceStartCmd start")
	defer func() {
		utils.LogTime("runServiceStartCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
			if exitCode == 0 {
				// there was an error and the exitcode
				// was not set to a non-zero value.
				// set it
				exitCode = 1
			}
		}
	}()

	ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, os.Kill)
	defer cancel()

	port := viper.GetInt(constants.ArgPort)
	if port < 1 || port > 65535 {
		panic("Invalid Port :: MUST be within range (1:65535)")
	}

	listen := db_local.StartListenType(viper.GetString(constants.ArgListenAddress))
	utils.FailOnError(listen.IsValid())

	invoker := constants.Invoker(cmdconfig.Viper().GetString(constants.ArgInvoker))
	utils.FailOnError(invoker.IsValid())

	err := db_local.EnsureDBInstalled(ctx)
	utils.FailOnError(err)

	// start db, refreshing connections
	startResult := db_local.StartServices(ctx, port, listen, invoker)
	utils.FailOnError(startResult.Error)

	if startResult.Status == db_local.ServiceFailedToStart {
		utils.ShowError(fmt.Errorf("steampipe service failed to start"))
		return
	}

	if startResult.Status == db_local.ServiceAlreadyRunning {
		if startResult.DbState.Invoker == constants.InvokerService {
			fmt.Println("Steampipe service is already running.")
			return
		}

		// check that we have the same port and listen parameters
		if port != startResult.DbState.Port {
			utils.FailOnError(fmt.Errorf("service is already running on port %d - cannot change port while it's running", startResult.DbState.Port))
		}
		if listen != startResult.DbState.ListenType {
			utils.FailOnError(fmt.Errorf("service is already running and listening on %s - cannot change listen type while it's running", startResult.DbState.ListenType))
		}

		// convert to being invoked by service
		startResult.DbState.Invoker = constants.InvokerService
		err = startResult.DbState.Save()
		if err != nil {
			utils.FailOnErrorWithMessage(err, "service was already running, but could not make it persistent")
		}
	}

	err = db_local.RefreshConnectionAndSearchPaths(ctx, invoker)
	if err != nil {
		db_local.StopServices(false, constants.InvokerService, nil)
		utils.FailOnError(err)
	}

	printStatus(startResult.DbState, startResult.PluginManagerState)

	if viper.GetBool(constants.ArgForeground) {
		runServiceInForeground(invoker)
	}
}

func runServiceInForeground(invoker constants.Invoker) {
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
			// if we have received this signal, then the user probably wants to shut down
			// everything. Shutdowns MUST NOT happen in cancellable contexts
			count, err := db_local.GetCountOfConnectedClients(context.Background())
			if err != nil {
				// report the error in the off chance that there's one
				utils.ShowError(err)
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

			db_local.StopServices(false, invoker, nil)
			fmt.Println("Steampipe service stopped.")
			return
		}
	}
}

func runServiceRestartCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("runServiceRestartCmd start")
	defer func() {
		utils.LogTime("runServiceRestartCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
			if exitCode == 0 {
				// there was an error and the exitcode
				// was not set to a non-zero value.
				// set it
				exitCode = 1
			}
		}
	}()

	// get current db statue
	currentDbState, err := db_local.GetState()
	utils.FailOnError(err)
	if currentDbState == nil {
		fmt.Println("Steampipe service is not running.")
		return
	}

	// stop db
	stopStatus, err := db_local.StopServices(viper.GetBool(constants.ArgForce), constants.InvokerService, nil)
	utils.FailOnErrorWithMessage(err, "could not stop current instance")
	if stopStatus != db_local.ServiceStopped {
		fmt.Println(`
Service stop failed.

Try using:
	steampipe service restart --force

to force a restart.
		`)
		return
	}

	// set the password in 'viper' so that it can be used by 'service start'
	viper.Set(constants.ArgServicePassword, currentDbState.Password)

	// start db
	startResult := db_local.StartServices(cmd.Context(), currentDbState.Port, currentDbState.ListenType, currentDbState.Invoker)
	utils.FailOnError(startResult.Error)
	if startResult.Status == db_local.ServiceFailedToStart {
		fmt.Println("Steampipe service was stopped, but failed to restart.")
		return
	}

	// refresh connections
	err = db_local.RefreshConnectionAndSearchPaths(cmd.Context(), constants.InvokerService)
	utils.FailOnError(err)
	fmt.Println("Steampipe service restarted.")

	printStatus(startResult.DbState, startResult.PluginManagerState)

}

func runServiceStatusCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("runServiceStatusCmd status")
	defer func() {
		utils.LogTime("runServiceStatusCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
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
		pmState, pmStateErr := plugin_manager.LoadPluginManagerState()

		if dbStateErr != nil || pmStateErr != nil {
			utils.ShowError(composeStateError(dbStateErr, pmStateErr))
			return
		}
		printStatus(dbState, pmState)
	}
}

func composeStateError(dbStateErr error, pmStateErr error) error {
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
	utils.LogTime("runServiceStopCmd stop")

	stoppedChan := make(chan bool, 1)
	var status db_local.StopStatus
	var err error
	var dbState *db_local.RunningDBInstanceInfo

	spinner := display.StartSpinnerAfterDelay("", constants.SpinnerShowTimeout, stoppedChan)

	defer func() {
		utils.LogTime("runServiceStopCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
			if exitCode == 0 {
				// there was an error and the exitcode
				// was not set to a non-zero value.
				// set it
				exitCode = 1
			}
		}
	}()

	force := cmdconfig.Viper().GetBool(constants.ArgForce)
	if force {
		status, err = db_local.StopServices(force, constants.InvokerService, spinner)
	} else {
		dbState, err = db_local.GetState()
		if err != nil {
			display.StopSpinner(spinner)
			utils.FailOnErrorWithMessage(err, "could not stop Steampipe service")
		}
		if dbState == nil {
			display.StopSpinner(spinner)
			fmt.Println("Steampipe service is not running.")
			return
		}
		if dbState.Invoker != constants.InvokerService {
			display.StopSpinner(spinner)
			printRunningImplicit(dbState.Invoker)
			return
		}

		// check if there are any connected clients to the service
		connectedClientCount, err := db_local.GetCountOfConnectedClients(cmd.Context())
		if err != nil {
			display.StopSpinner(spinner)
			utils.FailOnErrorWithMessage(err, "error during service stop")
		}

		if connectedClientCount > 0 {
			display.StopSpinner(spinner)
			printClientsConnected()
			return
		}

		status, _ = db_local.StopServices(false, constants.InvokerService, spinner)
	}

	if err != nil {
		display.StopSpinner(spinner)
		utils.ShowError(err)
		return
	}

	display.StopSpinner(spinner)

	switch status {
	case db_local.ServiceStopped:
		if dbState != nil {
			fmt.Printf("Steampipe service stopped [port %d].\n", dbState.Port)
		} else {
			fmt.Println("Steampipe service stopped.")
		}
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

	doneFetchingDetailsChan := make(chan bool)
	sp := display.StartSpinnerAfterDelay("Getting details", constants.SpinnerShowTimeout, doneFetchingDetailsChan)

	processes, err = db_local.FindAllSteampipePostgresInstances(ctx)
	close(doneFetchingDetailsChan)
	display.StopSpinner(sp)

	if err != nil {
		utils.ShowError(err)
		return
	}

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

	display.ShowWrappedTable(headers, rows, false)
}

func getServiceProcessDetails(process *psutils.Process) (string, string, string, db_local.StartListenType) {
	cmdLine, _ := process.CmdlineSlice()

	installDir := strings.TrimSuffix(cmdLine[0], db_local.ServiceExecutableRelativeLocation)
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

func printStatus(dbState *db_local.RunningDBInstanceInfo, pmState *plugin_manager.PluginManagerState) {
	if dbState == nil && !pmState.Running {
		fmt.Println("Service is not running")
		return
	}

	statusMessage := ""

	if dbState.Invoker == constants.InvokerService {
		msg := `
Steampipe service is running:

  Host(s):  %v
  Port:     %v
  Database: %v
  User:     %v
  Password: %v

Connection string:

  postgres://%v:%v@%v:%v/%v

Managing the Steampipe service:

  # Get status of the service
  steampipe service status

  # Restart the service
  steampipe service restart

  # Stop the service
  steampipe service stop
`
		statusMessage = fmt.Sprintf(msg, strings.Join(dbState.Listen, ", "), dbState.Port, dbState.Database, dbState.User, dbState.Password, dbState.User, dbState.Password, dbState.Listen[0], dbState.Port, dbState.Database)
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
		utils.ShowError(fmt.Errorf(`
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
