package cmd

import (
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
	"github.com/turbot/steampipe/db"
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
		AddStringFlag(constants.ArgListenAddress, "", string(db.ListenTypeNetwork), "Accept connections from: local (localhost only) or network (open)").
		AddStringFlag(constants.ArgListenAddressDeprecated, "", string(db.ListenTypeNetwork), "Accept connections from: local (localhost only) or network (open)", cmdconfig.FlagOptions.Deprecated(constants.ArgListenAddress)).
		// foreground enables the service to run in the foreground - till exit
		AddBoolFlag(constants.ArgForeground, "", false, "Run the service in the foreground").
		// Hidden flags for internal use
		AddStringFlag(constants.ArgInvoker, "", string(db.InvokerService), "Invoked by \"service\" or \"query\"", cmdconfig.FlagOptions.Hidden()).
		AddBoolFlag(constants.ArgRefresh, "", true, "Refresh connections on startup", cmdconfig.FlagOptions.Hidden())

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

	listen := db.StartListenType(cmdconfig.ListenAddress())
	if err := listen.IsValid(); err != nil {
		utils.ShowError(err)
		return
	}

	invoker := db.Invoker(cmdconfig.Viper().GetString(constants.ArgInvoker))
	if err := invoker.IsValid(); err != nil {
		utils.ShowError(err)
		return
	}

	db.EnsureDBInstalled()

	info, err := db.GetStatus()
	if err != nil {
		utils.ShowErrorWithMessage(err, "could not fetch service information")
		return
	}

	if info != nil {
		if info.Invoker == db.InvokerService {
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
		info.Invoker = db.InvokerService
		err = info.Save()
		if err != nil {
			utils.ShowErrorWithMessage(err, "service was already running, but could not make it persistent")
		}
	} else {
		status, err := db.StartDB(cmdconfig.DatabasePort(), listen, invoker, viper.GetBool(constants.ArgRefresh))
		if err != nil {
			utils.ShowError(err)
			return
		}

		if status == db.ServiceFailedToStart {
			utils.ShowError(fmt.Errorf("steampipe service failed to start"))
			return
		}

		if status == db.ServiceAlreadyRunning {
			utils.ShowError(fmt.Errorf("steampipe service is already running"))
			return
		}

		info, _ = db.GetStatus()
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
				newInfo, err := db.GetStatus()
				if err != nil {
					continue
				}
				if newInfo == nil {
					fmt.Println("Service stopped")
					return
				}
			case <-sigIntChannel:
				fmt.Print("\r")
				count, err := db.GetCountOfConnectedClients()
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
				db.StopDB(false, invoker)
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

	currentServiceStatus, err := db.GetStatus()

	if err != nil {
		utils.ShowError(errors.New("could not retrieve service status"))
		return
	}

	if currentServiceStatus == nil {
		fmt.Println("steampipe database service is not running")
		return
	}

	stopStatus, err := db.StopDB(viper.GetBool(constants.ArgForce), db.InvokerService)

	if err != nil {
		utils.ShowErrorWithMessage(err, "could not stop current instance")
		return
	}

	if stopStatus != db.ServiceStopped {
		fmt.Println(`
Service stop failed.

Try using:
	steampipe service restart --force
		
to force a restart.
		`)
		return
	}

	status, err := db.StartDB(currentServiceStatus.Port, currentServiceStatus.ListenType, currentServiceStatus.Invoker, viper.GetBool(constants.ArgRefresh))
	if err != nil {
		utils.ShowError(err)
		return
	}

	if status == db.ServiceFailedToStart {
		fmt.Println("Steampipe service was stopped, but failed to start")
		return
	}

	fmt.Println("Steampipe service restarted")

	if info, err := db.GetStatus(); err != nil {
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

	if !db.IsInstalled() {
		fmt.Println("Steampipe database service is NOT installed")
		return
	}
	if viper.GetBool(constants.ArgAll) {
		showAllStatus()
	} else {
		if info, err := db.GetStatus(); err != nil {
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

	processes, err = db.FindAllSteampipePostgresInstances()
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

func getServiceProcessDetails(process *psutils.Process) (string, string, string, db.StartListenType) {
	cmdLine, _ := process.CmdlineSlice()

	installDir := strings.TrimSuffix(cmdLine[0], db.ServiceExecutableRelativeLocation)
	var port string
	var listenType db.StartListenType

	for idx, param := range cmdLine {
		if param == "-p" {
			port = cmdLine[idx+1]
		}
		if strings.HasPrefix(param, "listen_addresses") {
			if strings.Contains(param, "localhost") {
				listenType = db.ListenTypeLocal
			} else {
				listenType = db.ListenTypeNetwork
			}
		}
	}

	return fmt.Sprintf("%d", process.Pid), installDir, port, listenType
}

func printStatus(info *db.RunningDBInstanceInfo) {

	statusMessage := ""

	if info.Invoker == db.InvokerService {
		msg := `
Steampipe database service is now running:

	Host(s):  %v
	Port:     %v
	Database: %v
	User:     %v
	Password: %v

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
		statusMessage = fmt.Sprintf(msg, strings.Join(info.Listen, ", "), info.Port, info.Database, info.User, info.Password, info.User, info.Password, info.Listen[0], info.Port, info.Database, db.SslMode())
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
	defer func() {
		utils.LogTime("runServiceStopCmd end")
		if r := recover(); r != nil {
			utils.ShowError(helpers.ToError(r))
		}
	}()

	var status db.StopStatus
	var err error

	force := cmdconfig.Viper().GetBool(constants.ArgForce)
	if force {
		status, err = db.StopDB(force, db.InvokerService)
	} else {
		info, err := db.GetStatus()
		if err != nil {
			utils.ShowErrorWithMessage(err, "could not stop service")
			return
		}
		if info == nil {
			fmt.Println("Service is not running")
			return
		}
		if info.Invoker != db.InvokerService {
			printRunningImplicit(info.Invoker)
			return
		}

		// check if there are any connected clients to the service
		connectedClientCount, err := db.GetCountOfConnectedClients()
		if err != nil {
			utils.ShowError(fmt.Errorf("error during stop"))
		}

		if connectedClientCount > 0 {
			printClientsConnected()
			return
		}

		status, err = db.StopDB(false, db.InvokerService)
	}

	if err != nil {
		utils.ShowError(err)
		return
	}

	switch status {
	case db.ServiceStopped:
		fmt.Println("Steampipe database service stopped")
	case db.ServiceNotRunning:
		fmt.Println("Service is not running")
	case db.ServiceStopFailed:
		fmt.Println("Could not stop service")
	case db.ServiceStopTimedOut:
		fmt.Println(`
Service stop operation timed-out.

This is probably because other clients are connected to the database service.

Disconnect all clients, or use	
	steampipe service stop --force

to force a shutdown
		`)

	}

}

func printRunningImplicit(invoker db.Invoker) {
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
