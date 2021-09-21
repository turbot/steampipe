package cmd

import (
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
	"github.com/turbot/steampipe/utils"
	"github.com/turbot/steampipe/workspace"
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
		// for now default port to -1 so we fall back to the default of the deprecated arg
		AddIntFlag(constants.ArgPort, "", constants.DatabaseDefaultPort, "Database service port.").
		AddIntFlag(constants.ArgPortDeprecated, "", constants.DatabaseDefaultPort, "Database service port.", cmdconfig.FlagOptions.Deprecated(constants.ArgPort)).
		// for now default listen address to empty so we fall back to the default of the deprecated arg
		AddStringFlag(constants.ArgListenAddress, "", string(db_local.ListenTypeNetwork), "Accept connections from: local (localhost only) or network (open)").
		AddStringFlag(constants.ArgListenAddressDeprecated, "", string(db_local.ListenTypeNetwork), "Accept connections from: local (localhost only) or network (open)", cmdconfig.FlagOptions.Deprecated(constants.ArgListenAddress)).
		AddStringFlag(constants.ArgServicePassword, "", "", "Set a service password for this session").
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

	port := cmdconfig.DatabasePort()
	if port < 1 || port > 65535 {
		panic("Invalid Port :: MUST be within range (1:65535)")
	}

	listen := db_local.StartListenType(cmdconfig.ListenAddress())
	utils.FailOnError(listen.IsValid())

	invoker := constants.Invoker(cmdconfig.Viper().GetString(constants.ArgInvoker))
	utils.FailOnError(invoker.IsValid())

	err := db_local.EnsureDBInstalled()
	utils.FailOnError(err)

	info, err := db_local.GetStatus()
	utils.FailOnErrorWithMessage(err, "could not fetch service information")

	if info != nil {
		if info.Invoker == constants.InvokerService {
			fmt.Println("Service is already running")
			return
		}

		// check that we have the same port and listen parameters
		if port != info.Port {
			utils.FailOnError(fmt.Errorf("service is already running on port %d - cannot change port while it's running", info.Port))
		}
		if listen != info.ListenType {
			utils.FailOnError(fmt.Errorf("service is already running and listening on %s - cannot change listen type while it's running", info.ListenType))
		}

		// convert
		info.Invoker = constants.InvokerService
		err = info.Save()
		if err != nil {
			utils.FailOnErrorWithMessage(err, "service was already running, but could not make it persistent")
		}
	} else {
		// start db, refreshing connections
		status, err := db_local.StartDB(cmdconfig.DatabasePort(), listen, invoker)
		utils.FailOnError(err)

		if status == db_local.ServiceFailedToStart {
			utils.ShowError(fmt.Errorf("steampipe service failed to start"))
			return
		}

		if status == db_local.ServiceAlreadyRunning {
			utils.FailOnError(fmt.Errorf("steampipe service is already running"))
		}

		err = db_local.RefreshConnectionAndSearchPaths(invoker)

		if err != nil {
			db_local.StopDB(false, constants.InvokerService, nil)
			utils.FailOnError(err)
		}
		info, _ = db_local.GetStatus()
	}
	printStatus(info)

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

	connectionWatcher, err := workspace.NewConnectionWatcher(invoker, func(error) {})
	utils.FailOnError(err)
	var lastCtrlC time.Time

	for {
		select {
		case <-checkTimer.C:
			// get the current status
			newInfo, err := db_local.GetStatus()
			if err != nil {
				continue
			}
			if newInfo == nil {
				fmt.Println("Service stopped")
				return
			}
		case <-sigIntChannel:
			fmt.Print("\r")
			count, err := db_local.GetCountOfConnectedClients()
			if err != nil {
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
			// close the connection watcher
			connectionWatcher.Close()
			fmt.Println("Stopping service")

			db_local.StopDB(false, invoker, nil)
			fmt.Println("Service Stopped")
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

	currentServiceStatus, err := db_local.GetStatus()

	utils.FailOnError(err)

	if currentServiceStatus == nil {
		fmt.Println("steampipe database service is not running")
		return
	}

	stopStatus, err := db_local.StopDB(viper.GetBool(constants.ArgForce), constants.InvokerService, nil)

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
	viper.Set(constants.ArgServicePassword, currentServiceStatus.Password)

	// start db, refreshing connections
	status, err := db_local.StartDB(currentServiceStatus.Port, currentServiceStatus.ListenType, currentServiceStatus.Invoker)
	if err != nil {
		utils.ShowError(err)
		return
	}

	if status == db_local.ServiceFailedToStart {
		fmt.Println("Steampipe service was stopped, but failed to start")
		return
	}

	err = db_local.RefreshConnectionAndSearchPaths(constants.InvokerService)
	utils.FailOnError(err)
	fmt.Println("Steampipe service restarted")

	if info, err := db_local.GetStatus(); err != nil {
		printStatus(info)
	}

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
		fmt.Println("Steampipe database service is NOT installed")
		return
	}
	if viper.GetBool(constants.ArgAll) {
		showAllStatus()
	} else {
		if info, err := db_local.GetStatus(); err != nil {
			utils.ShowError(fmt.Errorf("could not get Steampipe database service status"))
		} else if info != nil {
			printStatus(info)
		} else {
			fmt.Println("Steampipe database service is NOT running")
		}
	}
}

func runServiceStopCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("runServiceStopCmd stop")

	stoppedChan := make(chan bool, 1)
	var status db_local.StopStatus
	var err error
	var info *db_local.RunningDBInstanceInfo

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
		status, err = db_local.StopDB(force, constants.InvokerService, spinner)
	} else {
		info, err = db_local.GetStatus()
		if err != nil {
			display.StopSpinner(spinner)
			utils.FailOnErrorWithMessage(err, "could not stop service")
		}
		if info == nil {
			display.StopSpinner(spinner)
			fmt.Println("Service is not running")
			return
		}
		if info.Invoker != constants.InvokerService {
			display.StopSpinner(spinner)
			printRunningImplicit(info.Invoker)
			return
		}

		// check if there are any connected clients to the service
		connectedClientCount, err := db_local.GetCountOfConnectedClients()
		if err != nil {
			display.StopSpinner(spinner)
			utils.FailOnErrorWithMessage(err, "error during service stop")
		}

		if connectedClientCount > 0 {
			display.StopSpinner(spinner)
			printClientsConnected()
			return
		}

		status, _ = db_local.StopDB(false, constants.InvokerService, spinner)
	}

	if err != nil {
		display.StopSpinner(spinner)
		utils.ShowError(err)
		return
	}

	display.StopSpinner(spinner)

	switch status {
	case db_local.ServiceStopped:
		if info != nil {
			fmt.Printf("Steampipe database service stopped [port %d]\n", info.Port)
		} else {
			fmt.Println("Steampipe database service stopped")
		}
	case db_local.ServiceNotRunning:
		fmt.Println("Service is not running")
	case db_local.ServiceStopFailed:
		fmt.Println("Could not stop service")
	case db_local.ServiceStopTimedOut:
		fmt.Println(`
Service stop operation timed-out.

This is probably because other clients are connected to the database service.

Disconnect all clients, or use
	steampipe service stop --force

to force a shutdown
		`)

	}

}

func showAllStatus() {
	var processes []*psutils.Process
	var err error

	doneFetchingDetailsChan := make(chan bool)
	sp := display.StartSpinnerAfterDelay("Getting details", constants.SpinnerShowTimeout, doneFetchingDetailsChan)

	processes, err = db_local.FindAllSteampipePostgresInstances()
	close(doneFetchingDetailsChan)
	display.StopSpinner(sp)

	if err != nil {
		utils.ShowError(err)
		return
	}

	if len(processes) == 0 {
		fmt.Println("There are no steampipe services running")
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

func printStatus(info *db_local.RunningDBInstanceInfo) {

	statusMessage := ""

	if info.Invoker == constants.InvokerService {
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
		statusMessage = fmt.Sprintf(msg, strings.Join(info.Listen, ", "), info.Port, info.Database, info.User, info.Password, info.User, info.Password, info.Listen[0], info.Port, info.Database)
	} else {
		msg := `
Steampipe service was started for an active %s session. The service will exit when all active sessions exit.

To keep the service running after the %s session completes, use %s.
`

		statusMessage = fmt.Sprintf(
			msg,
			fmt.Sprintf("steampipe %s", info.Invoker),
			info.Invoker,
			constants.Bold("steampipe service start"),
		)
	}

	fmt.Println(statusMessage)
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

To force shutdown, press Ctrl+C again
	`
}
