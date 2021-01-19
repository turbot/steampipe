package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/turbot/steampipe/cmdconfig"

	"github.com/spf13/cobra"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/utils"
)

func init() {
	rootCmd.AddCommand(ServiceCmd())
}

// ServiceCmd :: Service management commands
func ServiceCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "service [command]",
		Args:  cobra.NoArgs,
		Short: "Steampipe service management",
		// TODO(nw) expand long description
		Long: `Steampipe service management.

Run Steampipe as a local service, exposing it as a database endpoint for
connection from any Postgres compatible database client.`,
	}

	cmd.AddCommand(ServiceStartCmd())
	cmd.AddCommand(ServiceStatusCmd())
	cmd.AddCommand(ServiceStopCmd())
	cmd.AddCommand(ServiceRestartCmd())

	return cmd
}

// ServiceStartCmd :: handler for service start
func ServiceStartCmd() *cobra.Command {
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
		AddBoolFlag("background", "", true, "Run service in the background").
		AddBoolFlag("refresh", "", true, "Refresh connections").
		// TODO(nw) default to the configuration option?
		AddIntFlag("db-port", "", constants.DatabasePort, "Database service port. Chooses a free port by default.").
		// TODO(nw) should be validated to an enumerated list
		AddStringFlag("listen", "", "network", "Accept connections from: local (localhost only) or network (open)")

	return cmd
}

// ServiceStatusCmd :: handler for service status
func ServiceStatusCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "status",
		Args:  cobra.NoArgs,
		Run:   runServiceStatusCmd,
		Short: "Status of the Steampipe service",
		Long: `Status of the Steampipe service.

Report current status of the Steampipe database service.`,
	}

	cmdconfig.OnCmd(cmd)

	return cmd
}

// ServiceStopCmd :: handler for service stop
func ServiceStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Args:  cobra.NoArgs,
		Run:   runServiceStopCmd,
		Short: "Stop Steampipe service",
		Long:  `Stop the Steampipe service.`,
	}

	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag(constants.ArgForce, "", false, "Forces the service to shutdown, releasing all open connections and ports")

	return cmd
}

// ServiceRestartCmd :: restarts the database service
func ServiceRestartCmd() *cobra.Command {
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
	logging.LogTime("runServiceStartCmd start")
	// 	// TODO(nw) - color me, replace hard-coding with variables / config

	if cmdconfig.Viper().GetInt("db-port") < 1 || cmdconfig.Viper().GetInt("db-port") > 65535 {
		fmt.Println("Invalid Port :: MUST be within range (1:65535)")
	}

	listen := db.StartListenType(cmdconfig.Viper().GetString("listen"))

	if err := listen.IsValid(); err != nil {
		utils.ShowError(err)
		return
	}

	db.EnsureDBInstalled()

	status, err := db.StartDB(cmdconfig.Viper().GetInt("db-port"), listen)
	if err != nil {
		utils.ShowError(err)
		return
	}

	if status == db.ServiceFailedToStart {
		fmt.Println("Steampipe Service failed to start")
		return
	}

	if status == db.ServiceAlreadyRunning {
		fmt.Println("Steampipe Service is already running")
		return
	}

	info, _ := db.GetStatus()

	printStatus(info)

	logging.LogTime("runServiceStartCmd end")
}

func runServiceRestartCmd(cmd *cobra.Command, args []string) {
	logging.LogTime("runServiceRestartCmd start")

	currentServiceStatus, err := db.GetStatus()

	if err != nil {
		utils.ShowError(errors.New("could not retrieve service status"))
		return
	}

	if currentServiceStatus == nil {
		fmt.Println("steampipe database service is not running")
		return
	}

	stopStatus, err := db.StopDB(cmdconfig.Viper().GetBool(constants.ArgForce))

	if err != nil {
		utils.ShowError(errors.New("could not stop current instance"))
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

	status, err := db.StartDB(currentServiceStatus.Port, currentServiceStatus.ListenType)
	if err != nil {
		utils.ShowError(err)
		return
	}

	if status == db.ServiceFailedToStart {
		fmt.Println("Steampipe Service was stopped, but failed to start")
		return
	}

	fmt.Println("Steampipe service restarted")

	if info, err := db.GetStatus(); err != nil {
		printStatus(info)
	}

	logging.LogTime("runServiceRestartCmd end")
}

func runServiceStatusCmd(cmd *cobra.Command, args []string) {
	logging.LogTime("runServiceStatusCmd status")

	if !db.IsInstalled() {
		fmt.Println("Steampipe database service is NOT installed")
	} else {
		if info, err := db.GetStatus(); err != nil {
			utils.ShowError(fmt.Errorf("Could not get Steampipe database service status"))
		} else if info != nil {
			printStatus(info)
		} else {
			fmt.Println("Steampipe database service is NOT running")
		}
	}

	logging.LogTime("runServiceStatusCmd end")
}

func printStatus(info *db.RunningDBInstanceInfo) {

	msg := `
Steampipe database service is now running:

  Host(s):  %v
  Port:     %v
  Database: %v
  User:     %v
  Password: %v

Connection string:

  postgres://%v:%v@%v:%v/%v?sslmode=disable

Steampipe service is running in the background.

  # Get status of the service
  steampipe service status
  
  # Restart the service
  steampipe service restart

  # Stop the service
  steampipe service stop

`

	fmt.Printf(msg, strings.Join(info.Listen, ", "), info.Port, info.Database, info.User, info.Password, info.User, info.Password, info.Listen[0], info.Port, info.Database)
}

func runServiceStopCmd(cmd *cobra.Command, args []string) {
	logging.LogTime("runServiceStopCmd stop")

	force := cmdconfig.Viper().GetBool(constants.ArgForce)
	status, err := db.StopDB(force)

	if err != nil {
		utils.ShowError(err)
	}

	switch status {
	case db.ServiceStopped:
		fmt.Println("Steampipe database service stopped")
	case db.ServiceStopFailed:
		fmt.Println("Could not stop service")
	case db.ServiceNotRunning:
		fmt.Println("Service is not running")
	case db.ServiceStopTimedOut:
		fmt.Println(`
Service stop operation timed-out.

This is probably because other clients are connected to the database service.

Disconnect all clients, or use	
	steampipe service stop --force

to force a shutdown
		`)

	}

	logging.LogTime("runServiceStopCmd end")
}
