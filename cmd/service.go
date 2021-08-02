package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/turbot/steampipe/db/local_db"

	psutils "github.com/shirou/gopsutil/process"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/display"
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
		AddStringFlag(constants.ArgListenAddress, "", string(local_db.ListenTypeNetwork), "Accept connections from: local (localhost only) or network (open)").
		AddStringFlag(constants.ArgListenAddressDeprecated, "", string(local_db.ListenTypeNetwork), "Accept connections from: local (localhost only) or network (open)", cmdconfig.FlagOptions.Deprecated(constants.ArgListenAddress)).
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
			exitCode = -1
		}
	}()

	port := cmdconfig.DatabasePort()
	if port < 1 || port > 65535 {
		fmt.Println("Invalid Port :: MUST be within range (1:65535)")
	}

	listen := local_db.StartListenType(cmdconfig.ListenAddress())
	if err := listen.IsValid(); err != nil {
		utils.ShowError(err)
		return
	}

	invoker := constants.Invoker(cmdconfig.Viper().GetString(constants.ArgInvoker))
	if err := invoker.IsValid(); err != nil {
		utils.ShowError(err)
		return
	}

	err := local_db.EnsureDBInstalled()
	if err != nil {
		utils.ShowError(err)
		return
	}

	info, err := local_db.GetStatus()
	if err != nil {
		utils.ShowErrorWithMessage(err, "could not fetch service information")
		return
	}

	if info != nil {
		if info.Invoker == constants.InvokerService {
			fmt.Println("Service is already running")
			return
		}

		// check that we have the same port and listen parameters
		if port != info.Port {
			utils.ShowError(fmt.Errorf("service is already running on port %d - cannot change port while it's running", info.Port))
			return
		}
		if listen != info.ListenType {
			utils.ShowError(fmt.Errorf("service is already running and listening on %s - cannot change listen type while it's running", info.ListenType))
			return
		}

		// convert
		info.Invoker = constants.InvokerService
		err = info.Save()
		if err != nil {
			utils.ShowErrorWithMessage(err, "service was already running, but could not make it persistent")
		}
	} else {
		// start db, refreshing connections
		status, err := local_db.StartDB(cmdconfig.DatabasePort(), listen, invoker)
		if err != nil {
			utils.ShowError(err)
			return
		}

		if status == local_db.ServiceFailedToStart {
			utils.ShowError(fmt.Errorf("steampipe service failed to start"))
			return
		}

		if status == local_db.ServiceAlreadyRunning {
			utils.ShowError(fmt.Errorf("steampipe service is already running"))
			return
		}
		if err := local_db.RefreshConnectionAndSearchPaths(invoker); err != nil {
			utils.ShowError(err)
			return
		}
		info, _ = local_db.GetStatus()
	}
	printStatus(info)

	if viper.GetBool(constants.ArgForeground) {
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
				newInfo, err := local_db.GetStatus()
				if err != nil {
					continue
				}
				if newInfo == nil {
					fmt.Println("Service stopped")
					return
				}
			case <-sigIntChannel:
				fmt.Print("\r")
				count, err := local_db.GetCountOfConnectedClients()
				if err != nil {
					return
				}
				if count > 0 {
					if lastCtrlC.IsZero() || time.Since(lastCtrlC) > 30*time.Second {
						lastCtrlC = time.Now()
						fmt.Println(buildForegroundClientsConnectedMsg())
						continue
					}
				}
				fmt.Println("Stopping service")
				local_db.StopDB(false, invoker, nil)
				fmt.Println("Service Stopped")
				return
			}
		}
	}
}

func runServiceRestartCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("runServiceRestartCmd start")
	defer func() {
		utils.LogTime("runServiceRestartCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	currentServiceStatus, err := local_db.GetStatus()

	if err != nil {
		utils.ShowError(errors.New("could not retrieve service status"))
		return
	}

	if currentServiceStatus == nil {
		fmt.Println("steampipe database service is not running")
		return
	}

	stopStatus, err := local_db.StopDB(viper.GetBool(constants.ArgForce), constants.InvokerService, nil)

	if err != nil {
		utils.ShowErrorWithMessage(err, "could not stop current instance")
		return
	}

	if stopStatus != local_db.ServiceStopped {
		fmt.Println(`
Service stop failed.

Try using:
	steampipe service restart --force
		
to force a restart.
		`)
		return
	}
	// start db, refreshing connections
	status, err := local_db.StartDB(currentServiceStatus.Port, currentServiceStatus.ListenType, currentServiceStatus.Invoker)
	if err != nil {
		utils.ShowError(err)
		return
	}

	if status == local_db.ServiceFailedToStart {
		fmt.Println("Steampipe service was stopped, but failed to start")
		return
	}

	if err := local_db.RefreshConnectionAndSearchPaths(constants.InvokerService); err != nil {
		utils.ShowError(err)
		return
	}

	fmt.Println("Steampipe service restarted")

	if info, err := local_db.GetStatus(); err != nil {
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

	if !local_db.IsInstalled() {
		fmt.Println("Steampipe database service is NOT installed")
		return
	}
	if viper.GetBool(constants.ArgAll) {
		showAllStatus()
	} else {
		if info, err := local_db.GetStatus(); err != nil {
			utils.ShowError(fmt.Errorf("could not get Steampipe database service status"))
		} else if info != nil {
			printStatus(info)
		} else {
			fmt.Println("Steampipe database service is NOT running")
		}
	}
}

func showAllStatus() {
	var processes []*psutils.Process
	var err error

	doneFetchingDetailsChan := make(chan bool)
	sp := display.StartSpinnerAfterDelay("Getting details", constants.SpinnerShowTimeout, doneFetchingDetailsChan)

	processes, err = local_db.FindAllSteampipePostgresInstances()
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

func getServiceProcessDetails(process *psutils.Process) (string, string, string, local_db.StartListenType) {
	cmdLine, _ := process.CmdlineSlice()

	installDir := strings.TrimSuffix(cmdLine[0], local_db.ServiceExecutableRelativeLocation)
	var port string
	var listenType local_db.StartListenType

	for idx, param := range cmdLine {
		if param == "-p" {
			port = cmdLine[idx+1]
		}
		if strings.HasPrefix(param, "listen_addresses") {
			if strings.Contains(param, "localhost") {
				listenType = local_db.ListenTypeLocal
			} else {
				listenType = local_db.ListenTypeNetwork
			}
		}
	}

	return fmt.Sprintf("%d", process.Pid), installDir, port, listenType
}

func printStatus(info *local_db.RunningDBInstanceInfo) {

	statusMessage := ""

	if info.Invoker == constants.InvokerService {
		msg := `
Steampipe database service is now running:

	Host(s):  %v
	Port:     %v
	Database: %v
	User:     %v
	Password: %v
	SSL:      %v

Connection string:

	postgres://%v:%v@%v:%v/%v?sslmode=%v

Managing Steampipe service:

	# Get status of the service
	steampipe service status
	
	# Restart the service
	steampipe service restart

	# Stop the service
	steampipe service stop
	
`
		statusMessage = fmt.Sprintf(msg, strings.Join(info.Listen, ", "), info.Port, info.Database, info.User, info.Password, local_db.SslStatus(), info.User, info.Password, info.Listen[0], info.Port, info.Database, local_db.SslMode())
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

func runServiceStopCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("runServiceStopCmd stop")

	stoppedChan := make(chan bool, 1)
	var status local_db.StopStatus
	var err error
	var info *local_db.RunningDBInstanceInfo

	spinner := display.StartSpinnerAfterDelay("", constants.SpinnerShowTimeout, stoppedChan)

	defer func() {
		utils.LogTime("runServiceStopCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	force := cmdconfig.Viper().GetBool(constants.ArgForce)
	if force {
		status, err = local_db.StopDB(force, constants.InvokerService, spinner)
	} else {
		info, err = local_db.GetStatus()
		if err != nil {
			display.StopSpinner(spinner)
			utils.ShowErrorWithMessage(err, "could not stop service")
			return
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
		connectedClientCount, err := local_db.GetCountOfConnectedClients()
		if err != nil {
			display.StopSpinner(spinner)
			utils.ShowError(utils.PrefixError(err, "error during service stop"))
		}

		if connectedClientCount > 0 {
			display.StopSpinner(spinner)
			printClientsConnected()
			return
		}

		status, _ = local_db.StopDB(false, constants.InvokerService, spinner)
	}

	if err != nil {
		display.StopSpinner(spinner)
		utils.ShowError(err)
		return
	}

	display.StopSpinner(spinner)

	switch status {
	case local_db.ServiceStopped:
		fmt.Println("Steampipe database service stopped")
	case local_db.ServiceNotRunning:
		fmt.Println("Service is not running")
	case local_db.ServiceStopFailed:
		fmt.Println("Could not stop service")
	case local_db.ServiceStopTimedOut:
		fmt.Println(`
Service stop operation timed-out.

This is probably because other clients are connected to the database service.

Disconnect all clients, or use	
	steampipe service stop --force

to force a shutdown
		`)

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

To force shutdown, press Ctrl+C again
	`
}
