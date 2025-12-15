package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thediveo/enumflag/v2"
	"github.com/turbot/go-kit/helpers"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/contexthelpers"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/v2/pkg/cmdconfig"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/query"
	"github.com/turbot/steampipe/v2/pkg/query/queryexecute"
	"github.com/turbot/steampipe/v2/pkg/statushooks"
)

// variable used to assign the timing mode flag
var queryTimingMode = constants.QueryTimingModeOff

// variable used to assign the output mode flag
var queryOutputMode = constants.QueryOutputModeTable

// queryConfig holds the configuration needed for query validation
// This avoids concurrent access to global viper state
type queryConfig struct {
	snapshot bool
	share    bool
	export   []string
	output   string
}

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
	}

	// Notes:
	// * In the future we may add --csv and --json flags as shortcuts for --output
	cmdconfig.
		OnCmd(cmd).
		AddCloudFlags().
		AddWorkspaceDatabaseFlag().
		AddBoolFlag(pconstants.ArgHelp, false, "Help for query", cmdconfig.FlagOptions.WithShortHand("h")).
		AddBoolFlag(pconstants.ArgHeader, true, "Include column headers csv and table output").
		AddStringFlag(pconstants.ArgSeparator, ",", "Separator string for csv output").
		AddVarFlag(enumflag.New(&queryOutputMode, pconstants.ArgOutput, constants.QueryOutputModeIds, enumflag.EnumCaseInsensitive),
			pconstants.ArgOutput,
			fmt.Sprintf("Output format; one of: %s", strings.Join(constants.FlagValues(constants.QueryOutputModeIds), ", "))).
		AddVarFlag(enumflag.New(&queryTimingMode, pconstants.ArgTiming, constants.QueryTimingModeIds, enumflag.EnumCaseInsensitive),
			pconstants.ArgTiming,
			fmt.Sprintf("Display query timing; one of: %s", strings.Join(constants.FlagValues(constants.QueryTimingModeIds), ", ")),
			cmdconfig.FlagOptions.NoOptDefVal(pconstants.ArgOn)).
		AddStringSliceFlag(pconstants.ArgSearchPath, nil, "Set a custom search_path for the steampipe user for a query session (comma-separated)").
		AddStringSliceFlag(pconstants.ArgSearchPathPrefix, nil, "Set a prefix to the current search path for a query session (comma-separated)").
		AddBoolFlag(pconstants.ArgInput, true, "Enable interactive prompts").
		AddBoolFlag(pconstants.ArgSnapshot, false, "Create snapshot in Turbot Pipes with the default (workspace) visibility").
		AddBoolFlag(pconstants.ArgShare, false, "Create snapshot in Turbot Pipes with 'anyone_with_link' visibility").
		AddStringArrayFlag(pconstants.ArgSnapshotTag, nil, "Specify tags to set on the snapshot").
		AddStringFlag(pconstants.ArgSnapshotTitle, "", "The title to give a snapshot").
		AddIntFlag(pconstants.ArgDatabaseQueryTimeout, 0, "The query timeout").
		AddStringSliceFlag(pconstants.ArgExport, nil, "Export output to file, supported format: sps (snapshot)").
		AddStringFlag(pconstants.ArgSnapshotLocation, "", "The location to write snapshots - either a local file path or a Turbot Pipes workspace").
		AddBoolFlag(pconstants.ArgProgress, true, "Display snapshot upload status")

	return cmd
}

func runQueryCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	utils.LogTime("cmd.runQueryCmd start")
	defer func() {
		utils.LogTime("cmd.runQueryCmd end")
		if r := recover(); r != nil {
			error_helpers.ShowError(ctx, helpers.ToError(r))
		}
	}()

	// Read configuration from viper once to avoid concurrent access issues
	cfg := &queryConfig{
		snapshot: viper.IsSet(pconstants.ArgSnapshot),
		share:    viper.IsSet(pconstants.ArgShare),
		export:   viper.GetStringSlice(pconstants.ArgExport),
		output:   viper.GetString(pconstants.ArgOutput),
	}

	// validate args
	err := validateQueryArgs(ctx, args, cfg)
	error_helpers.FailOnError(err)

	// if diagnostic mode is set, print out config and return
	if _, ok := os.LookupEnv(constants.EnvConfigDump); ok {
		cmdconfig.DisplayConfig()
		return
	}

	if len(args) == 0 {
		// no positional arguments - check if there's anything on stdin
		if stdinData := getPipedStdinData(); len(stdinData) > 0 {
			// we have data - treat this as an argument
			args = append(args, stdinData)
		}
	}

	// enable paging only in interactive mode
	interactiveMode := len(args) == 0
	// set config to indicate whether we are running an interactive query
	viper.Set(constants.ConfigKeyInteractive, interactiveMode)

	// initialize the cancel handler - for context cancellation
	initCtx, cancel := context.WithCancel(ctx)
	contexthelpers.StartCancelHandler(cancel)

	// start the initializer
	initData := query.NewInitData(initCtx, args)
	if initData.Result.Error != nil {
		exitCode = constants.ExitCodeInitializationFailed
		error_helpers.ShowError(ctx, initData.Result.Error)
		return
	}
	defer initData.Cleanup(ctx)

	var failures int
	switch {
	case interactiveMode:
		err = queryexecute.RunInteractiveSession(ctx, initData)
	default:
		// NOTE: disable any status updates - we do not want 'loading' output from any queries
		ctx = statushooks.DisableStatusHooks(ctx)

		// fall through to running a batch query
		failures, err = queryexecute.RunBatchSession(ctx, initData)
	}

	// check for err and set the exit code else set the exit code if some queries failed or some rows returned an error
	if err != nil {
		exitCode = constants.ExitCodeInitializationFailed
		error_helpers.ShowError(ctx, err)
	} else if failures > 0 {
		exitCode = constants.ExitCodeQueryExecutionFailed
	}
}

func validateQueryArgs(ctx context.Context, args []string, cfg *queryConfig) error {
	interactiveMode := len(args) == 0
	if interactiveMode && (cfg.snapshot || cfg.share) {
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		return sperr.New("cannot share snapshots in interactive mode")
	}
	if interactiveMode && len(cfg.export) > 0 {
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		return sperr.New("cannot export query results in interactive mode")
	}
	// if share or snapshot args are set, there must be a query specified
	err := cmdconfig.ValidateSnapshotArgs(ctx)
	if err != nil {
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		return err
	}

	validOutputFormats := []string{constants.OutputFormatLine, constants.OutputFormatCSV, constants.OutputFormatTable, constants.OutputFormatJSON, constants.OutputFormatSnapshot, constants.OutputFormatSnapshotShort, constants.OutputFormatNone}
	if !slices.Contains(validOutputFormats, cfg.output) {
		exitCode = constants.ExitCodeInsufficientOrWrongInputs
		return sperr.New("invalid output format: '%s', must be one of [%s]", cfg.output, strings.Join(validOutputFormats, ", "))
	}

	return nil
}

// getPipedStdinData reads the Standard Input and returns the available data as a string
// if and only if the data was piped to the process
func getPipedStdinData() string {
	fi, err := os.Stdin.Stat()
	if err != nil {
		error_helpers.ShowWarning("could not fetch information about STDIN")
		return ""
	}
	if (fi.Mode()&os.ModeCharDevice) == 0 && fi.Size() > 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			error_helpers.ShowWarning("could not read from STDIN")
			return ""
		}
		return string(data)
	}
	return ""
}
