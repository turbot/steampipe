package control

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe/pkg/control/controldisplay"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/statushooks"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/initialisation"
	"github.com/turbot/steampipe/pkg/workspace"
)

type InitData struct {
	initialisation.InitData
	OutputFormatter controldisplay.Formatter
}

// NewInitData returns a new InitData object
// It also starts an asynchronous population of the object
// InitData.Done closes after asynchronous initialization completes
func NewInitData(ctx context.Context) *InitData {

	statushooks.SetStatus(ctx, "Initializing...")
	defer statushooks.Done(ctx)

	// load the workspace
	w, err := workspace.LoadWorkspacePromptingForVariables(ctx)
	if err != nil {
		return &InitData{
			InitData: *initialisation.NewErrorInitData(fmt.Errorf("failed to load workspace: %s", err.Error())),
		}
	}

	i := &InitData{
		InitData: *initialisation.NewInitData(w).Init(ctx, constants.InvokerCheck),
	}
	if i.Result.Error != nil {
		return i
	}

	if !w.ModfileExists() {
		i.Result.Error = workspace.ErrorNoModDefinition
	}

	if viper.GetString(constants.ArgOutput) == constants.OutputFormatNone {
		// set progress to false
		viper.Set(constants.ArgProgress, false)
	}
	// set color schema
	err = initialiseCheckColorScheme()
	if err != nil {
		i.Result.Error = err
		return i
	}

	if len(i.Workspace.GetResourceMaps().Controls) == 0 {
		i.Result.AddWarnings("no controls found in current workspace")
	}

	if err := controldisplay.EnsureTemplates(); err != nil {
		i.Result.Error = err
		return i
	}

	if len(viper.GetStringSlice(constants.ArgExport)) > 0 {
		i.registerCheckExporters()
		// validate required export formats
		if err := i.ExportManager.ValidateExportFormat(viper.GetStringSlice(constants.ArgExport)); err != nil {
			i.Result.Error = err
			return i
		}
	}

	output := viper.GetString(constants.ArgOutput)
	formatter, err := parseOutputArg(output)
	if err != nil {
		i.Result.Error = err
		return i
	}
	i.OutputFormatter = formatter

	return i
}

// register exporters for each of the supported check formats
func (initData *InitData) registerCheckExporters() {
	exporters, err := controldisplay.GetExporters()
	error_helpers.FailOnErrorWithMessage(err, "failed to load exporters")

	// register all exporters
	initData.RegisterExporters(exporters...)
}

// parseOutputArg parses the --output flag value and returns the Formatter that can format the data
func parseOutputArg(arg string) (formatter controldisplay.Formatter, err error) {
	formatResolver, err := controldisplay.NewFormatResolver()
	if err != nil {
		return nil, err
	}

	return formatResolver.GetFormatter(arg)
}

func initialiseCheckColorScheme() error {
	theme := viper.GetString(constants.ArgTheme)
	if !viper.GetBool(constants.ConfigKeyIsTerminalTTY) {
		// enforce plain output for non-terminals
		theme = "plain"
	}
	themeDef, ok := controldisplay.ColorSchemes[theme]
	if !ok {
		return fmt.Errorf("invalid theme '%s'", theme)
	}
	scheme, err := controldisplay.NewControlColorScheme(themeDef)
	if err != nil {
		return err
	}
	controldisplay.ControlColors = scheme
	return nil
}
