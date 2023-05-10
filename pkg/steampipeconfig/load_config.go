package steampipeconfig

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/hashicorp/hcl/v2"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
	"github.com/turbot/steampipe/pkg/utils"
	"github.com/turbot/steampipe/sperr"
)

var GlobalConfig *SteampipeConfig
var defaultConfigFileName = "default.spc"

// LoadSteampipeConfig loads the HCL connection config and workspace options
func LoadSteampipeConfig(modLocation string, commandName string) (*SteampipeConfig, *modconfig.ErrorAndWarnings) {
	utils.LogTime("steampipeconfig.LoadSteampipeConfig start")
	defer utils.LogTime("steampipeconfig.LoadSteampipeConfig end")

	if err := ensureDefaultConfigFile(filepaths.EnsureConfigDir()); err != nil {
		return nil, modconfig.NewErrorsAndWarning(
			sperr.WrapWithMessage(
				err,
				"could not create default config",
			),
		)
	}
	return loadSteampipeConfig(modLocation, commandName)
}

// LoadConnectionConfig loads the connection config but not the workspace options
// this is called by the fdw
func LoadConnectionConfig() (*SteampipeConfig, *modconfig.ErrorAndWarnings) {
	return LoadSteampipeConfig("", "")
}

func ensureDefaultConfigFile(configFolder string) error {
	defaultConfigFile := filepath.Join(configFolder, defaultConfigFileName)
	if _, err := os.Stat(defaultConfigFile); os.IsNotExist(err) {
		err = os.WriteFile(defaultConfigFile, []byte(constants.DefaultConnectionConfigContent), 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadSteampipeConfig(modLocation string, commandName string) (steampipeConfig *SteampipeConfig, errorsAndWarnings *modconfig.ErrorAndWarnings) {
	utils.LogTime("steampipeconfig.loadSteampipeConfig start")
	defer utils.LogTime("steampipeconfig.loadSteampipeConfig end")

	errorsAndWarnings = modconfig.NewErrorsAndWarning(nil)
	defer func() {
		if r := recover(); r != nil {
			errorsAndWarnings = modconfig.NewErrorsAndWarning(helpers.ToError(r))
		}
	}()

	steampipeConfig = NewSteampipeConfig(commandName)

	// load config from the installation folder -  load all spc files from config directory
	include := filehelpers.InclusionsFromExtensions(constants.ConnectionConfigExtensions)
	loadOptions := &loadConfigOptions{include: include}
	if ew := loadConfig(filepaths.EnsureConfigDir(), steampipeConfig, loadOptions); ew != nil {
		if ew.GetError() != nil {
			return nil, ew.WrapErrorWithMessage("failed to load config")
		}
		// merge the warning from this call
		errorsAndWarnings.AddWarning(ew.Warnings...)
	}

	// now load config from the workspace folder, if provided
	// this has precedence and so will overwrite any config which has already been set
	// check workspace folder exists
	if modLocation != "" {
		if _, err := os.Stat(modLocation); os.IsNotExist(err) {
			return nil, modconfig.NewErrorsAndWarning(fmt.Errorf("mod location '%s' does not exist", modLocation))
		}

		// only include workspace.spc from workspace directory
		include = filehelpers.InclusionsFromFiles([]string{filepaths.WorkspaceConfigFileName})
		// update load options to ONLY allow terminal options
		loadOptions = &loadConfigOptions{include: include, allowedOptions: []string{options.TerminalBlock}}
		if ew := loadConfig(modLocation, steampipeConfig, loadOptions); ew != nil {
			if ew.GetError() != nil {
				return nil, ew.WrapErrorWithMessage("failed to load workspace config")
			}

			// merge the warning from this call
			errorsAndWarnings.AddWarning(ew.Warnings...)
		}
	}

	// now set default options on all connections without options set
	// this is needed as the connection config is also loaded by the FDW which has no access to viper
	steampipeConfig.setDefaultConnectionOptions()

	// now validate the config
	warnings, errors := steampipeConfig.Validate()
	logValidationResult(warnings, errors)

	return steampipeConfig, errorsAndWarnings
}

func logValidationResult(warnings []string, errors []string) {
	if len(warnings) > 0 {
		error_helpers.ShowWarning(buildValidationLogString(warnings, "warning"))
		log.Printf("[TRACE] %s", buildValidationLogString(warnings, "warning"))
	}
	if len(errors) > 0 {
		error_helpers.ShowWarning(buildValidationLogString(errors, "error"))
		log.Printf("[TRACE] %s", buildValidationLogString(errors, "error"))
	}
}

func buildValidationLogString(items []string, validationType string) string {
	count := len(items)
	if count == 0 {
		return ""
	}
	var str strings.Builder
	str.WriteString(fmt.Sprintf("connection config has has %d validation %s:\n",
		count,
		pluralize.NewClient().Pluralize(validationType, count, false),
	))
	for _, w := range items {
		str.WriteString(fmt.Sprintf("\t %s\n", w))
	}
	return str.String()
}

// load config from the given folder and update steampipeConfig
// NOTE: this mutates steampipe config
type loadConfigOptions struct {
	include        []string
	allowedOptions []string
}

func loadConfig(configFolder string, steampipeConfig *SteampipeConfig, opts *loadConfigOptions) *modconfig.ErrorAndWarnings {
	// get all the config files in the directory
	configPaths, err := filehelpers.ListFiles(configFolder, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: opts.include,
	})

	if err != nil {
		log.Printf("[WARN] loadConfig: failed to get config file paths: %v\n", err)
		return modconfig.NewErrorsAndWarning(err)
	}
	if len(configPaths) == 0 {
		return nil
	}

	fileData, diags := parse.LoadFileData(configPaths...)
	if diags.HasErrors() {
		log.Printf("[WARN] loadConfig: failed to load all config files: %v\n", err)
		return modconfig.DiagsToErrorsAndWarnings("Failed to load all config files", diags)
	}

	body, diags := parse.ParseHclFiles(fileData)
	if diags.HasErrors() {
		return modconfig.DiagsToErrorsAndWarnings("Failed to load all config files", diags)
	}

	// do a partial decode
	content, moreDiags := body.Content(parse.ConfigBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return modconfig.DiagsToErrorsAndWarnings("Failed to load config", diags)
	}

	// store block types which we have found in this folder - each is only allowed once
	// NOTE this is different to merging options with options already populated in the passed-in steampipe config
	// this is valid because the same block may be defined in the config folder and the workspace
	optionBlockMap := map[string]bool{}

	for _, block := range content.Blocks {
		switch block.Type {
		case modconfig.BlockTypeConnection:
			connection, moreDiags := parse.DecodeConnection(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			_, alreadyThere := steampipeConfig.Connections[connection.Name]
			if alreadyThere {
				return modconfig.NewErrorsAndWarning(sperr.New("duplicate connection name: '%s' in '%s'", connection.Name, block.TypeRange.Filename))
			}
			if ok, errorMessage := db_common.IsSchemaNameValid(connection.Name); !ok {
				return modconfig.NewErrorsAndWarning(sperr.New("invalid connection name: '%s' in '%s'. %s ", connection.Name, block.TypeRange.Filename, errorMessage))
			}
			steampipeConfig.Connections[connection.Name] = connection

		case modconfig.BlockTypeOptions:
			// check this options type is permitted based on the options passed in
			if err := optionsBlockPermitted(block, optionBlockMap, opts); err != nil {
				return modconfig.NewErrorsAndWarning(err)
			}
			opts, moreDiags := parse.DecodeOptions(block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				continue
			}
			// set options on steampipe config
			// if options are already set, this will merge the new options over the top of the existing options
			// i.e. new options have precedence
			e := steampipeConfig.SetOptions(opts)
			if e.GetError() != nil {
				// we should never get an error here, since SetOptions
				// only sets warnings
				// putting this here only for good-practice
				return e
			}
			if len(e.Warnings) > 0 {
				for _, warning := range e.Warnings {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagWarning,
						Summary:  warning,
						Subject:  &block.DefRange,
					})
				}
			}
		}
	}

	if diags.HasErrors() {
		return modconfig.DiagsToErrorsAndWarnings("Failed to load config", diags)
	}

	return modconfig.DiagsToErrorsAndWarnings("", diags)
}

func optionsBlockPermitted(block *hcl.Block, blockMap map[string]bool, opts *loadConfigOptions) error {
	// keep track of duplicate block types
	blockType := block.Labels[0]
	if _, ok := blockMap[blockType]; ok {
		return fmt.Errorf("multiple instances of '%s' options block", blockType)
	}
	blockMap[blockType] = true
	permitted := len(opts.allowedOptions) == 0 ||
		helpers.StringSliceContains(opts.allowedOptions, blockType)

	if !permitted {
		return fmt.Errorf("'%s' options block is not permitted", blockType)
	}
	return nil
}
