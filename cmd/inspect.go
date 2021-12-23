package cmd

import (
	"context"
	"log"
	"runtime/debug"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_client"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/db/db_local"
	"github.com/turbot/steampipe/display"
	"github.com/turbot/steampipe/inspector"
	"github.com/turbot/steampipe/utils"
)

func inspectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "inspect",
		TraverseChildren: true,
		Args:             cobra.ArbitraryArgs,
		Run:              runInspectCmd,
		Short:            "",
		Long:             ``,
	}

	cmdconfig.
		OnCmd(cmd)
	return cmd
}

func runInspectCmd(cmd *cobra.Command, args []string) {
	utils.LogTime("cmd.runInspectCmd start")
	var client db_common.Client
	var err error
	var spinner *spinner.Spinner

	defer func() {
		utils.LogTime("cmd.runInspectCmd end")
		if r := recover(); r != nil {
			debug.PrintStack()
			utils.ShowError(helpers.ToError(r))
			err = helpers.ToError(r)
		}
		if client != nil {
			log.Printf("[TRACE] close client")
			client.Close()
		}
		if err != nil {
			exitCode = -1
		}
		display.StopSpinner(spinner)
	}()

	spinner = display.ShowSpinner("Initializing...")

	client, err = initializeInspect(cmd.Context(), spinner)
	utils.FailOnError(err)

	schemaMetadata, err := client.GetSchemaFromDB(cmd.Context(), client.ForeignSchemas())
	utils.FailOnError(err)

	connectionMap := *client.ConnectionMap()

	display.StopSpinner(spinner)

	if len(args) == 0 {
		utils.FailOnError(inspector.ListConnections(cmd.Context(), *schemaMetadata, connectionMap))
		return
	}

	if err := inspector.DescribeConnection(cmd.Context(), args[0], *schemaMetadata, connectionMap); err == nil {
		// this is a valid connection name
		return
	}

	if err := inspector.DescribeTable(cmd.Context()); err == nil {
		// this is a valid connection name
		return
	}

	return
}

func initializeInspect(ctx context.Context, spinner *spinner.Spinner) (db_common.Client, error) {
	display.UpdateSpinnerMessage(spinner, "Connecting to service...")
	// get a client
	var client db_common.Client
	var err error
	if connectionString := viper.GetString(constants.ArgConnectionString); connectionString != "" {
		client, err = db_client.NewDbClient(ctx, connectionString)
	} else {
		// stop the spinner
		display.StopSpinner(spinner)
		// when starting the database, installers may trigger their own spinners
		client, err = db_local.GetLocalClient(ctx, constants.InvokerInspect)
		// resume the spinner
		display.ResumeSpinner(spinner)
	}

	client.RefreshConnectionAndSearchPaths(ctx)

	return client, err
}
