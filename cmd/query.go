package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/interactive"
	"github.com/turbot/steampipe/pkg/query"
	"github.com/turbot/steampipe/pkg/query/queryexecute"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/pkg/workspace"
)

func queryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "query",
		TraverseChildren: true,
		Args:             cobra.ArbitraryArgs,
		Run:              runQueryCmd,
		Short:            "Execute SQL queries interactively or by argument",
		Long: `Execute SQL queries interactively, or by a query argument.

Open a interactive SQL query console to Steampipe to explore your data and run
multiple queries. If QUERY is passed on the command line then it will be run
immediately and the command will exit.

Examples:

  # Open an interactive query console
  steampipe query

  # Run a specific query directly
  steampipe query "select * from cloud"`,

		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			workspace, err := workspace.LoadResourceNames(viper.GetString(constants.ArgWorkspaceChDir))
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveError
			}
			namedQueries := []string{}
			for _, name := range workspace.GetSortedNamedQueryNames() {
				if strings.HasPrefix(name, toComplete) {
					namedQueries = append(namedQueries, name)
				}
			}
			return namedQueries, cobra.ShellCompDirectiveNoFileComp
		},
	}

	// Notes:
	// * In the future we may add --csv and --json flags as shortcuts for --output
	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag(constants.ArgHelp, "h", false, "Help for query").
		AddBoolFlag(constants.ArgHeader, "", true, "Include column headers csv and table output").
		AddStringFlag(constants.ArgSeparator, "", ",", "Separator string for csv output").
		AddStringFlag(constants.ArgOutput, "", "table", "Output format: line, csv, json or table").
		AddBoolFlag(constants.ArgTiming, "", false, "Turn on the timer which reports query time.").
		AddBoolFlag(constants.ArgWatch, "", true, "Watch SQL files in the current workspace (works only in interactive mode)").
		AddStringSliceFlag(constants.ArgSearchPath, "", nil, "Set a custom search_path for the steampipe user for a query session (comma-separated)").
		AddStringSliceFlag(constants.ArgSearchPathPrefix, "", nil, "Set a prefix to the current search path for a query session (comma-separated)").
		AddStringSliceFlag(constants.ArgVarFile, "", nil, "Specify a file containing variable values").
		// NOTE: use StringArrayFlag for ArgVariable, not StringSliceFlag
		// Cobra will interpret values passed to a StringSliceFlag as CSV,
		// where args passed to StringArrayFlag are not parsed and used raw
		AddStringArrayFlag(constants.ArgVariable, "", nil, "Specify the value of a variable").
		AddBoolFlag(constants.ArgInput, "", true, "Enable interactive prompts").
		AddStringFlag(constants.ArgSnapshot, "", "", "Create snapshot in Steampipe Cloud with the default (workspace) visibility.", cmdconfig.FlagOptions.NoOptDefVal(constants.ArgShareNoOptDefault)).
		AddStringFlag(constants.ArgShare, "", "", "Create snapshot in Steampipe Cloud with 'anyone_with_link' visibility.", cmdconfig.FlagOptions.NoOptDefVal(constants.ArgShareNoOptDefault))

	return cmd
}

func runQueryCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("cmd.runQueryCmd start")
	defer func() {
		utils.LogTime("cmd.runQueryCmd end")
		if r := recover(); r != nil {
			utils.ShowError(ctx, helpers.ToError(r))
		}
	}()

	if stdinData := getPipedStdinData(); len(stdinData) > 0 {
		args = append(args, stdinData)
	}

	// validate args
	utils.FailOnError(validateQueryArgs())

	cloudMetadata, err := cmdconfig.GetCloudMetadata()
	utils.FailOnError(err)

	// enable spinner only in interactive mode
	interactiveMode := len(args) == 0
	// set config to indicate whether we are running an interactive query
	viper.Set(constants.ConfigKeyInteractive, interactiveMode)

	// load the workspace
	w, err := interactive.LoadWorkspacePromptingForVariables(ctx)
	utils.FailOnErrorWithMessage(err, "failed to load workspace")

	// set cloud metadata (may be nil)
	w.CloudMetadata = cloudMetadata

	// so we have loaded a workspace - be sure to close it
	defer w.Close()

	// start the initializer
	initData := query.NewInitData(ctx, w, args)

	if interactiveMode {
		queryexecute.RunInteractiveSession(ctx, initData)
	} else {
		// NOTE: disable any status updates - we do not want 'loading' output from any queries
		ctx = statushooks.DisableStatusHooks(ctx)
		// set global exit code
		exitCode = queryexecute.RunBatchSession(ctx, initData)
	}
}

func validateQueryArgs() error {
	// only 1 of 'share' and 'snapshot' may be set
	if len(viper.GetString(constants.ArgShare)) > 0 && len(viper.GetString(constants.ArgShare)) > 0 {
		return fmt.Errorf("only 1 of 'share' and 'dashboard' may be set")
	}
	return nil
}

// getPipedStdinData reads the Standard Input and returns the available data as a string
// if and only if the data was piped to the process
func getPipedStdinData() string {
	fi, err := os.Stdin.Stat()
	if err != nil {
		utils.ShowWarning("could not fetch information about STDIN")
		return ""
	}
	stdinData := ""
	if (fi.Mode()&os.ModeCharDevice) == 0 && fi.Size() > 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			stdinData = fmt.Sprintf("%s%s", stdinData, scanner.Text())
		}
	}
	return stdinData
}
