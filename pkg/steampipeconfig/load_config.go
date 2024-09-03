package steampipeconfig

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gertd/go-pluralize"
	"github.com/hashicorp/hcl/v2"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	pconstants "github.com/turbot/pipe-fittings/constants"
	perror_helpers "github.com/turbot/pipe-fittings/error_helpers"
	pfilepaths "github.com/turbot/pipe-fittings/filepaths"
	"github.com/turbot/pipe-fittings/hclhelpers"
	"github.com/turbot/pipe-fittings/ociinstaller/versionfile"
	"github.com/turbot/pipe-fittings/utils"
	"github.com/turbot/steampipe-plugin-sdk/v5/sperr"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
)

var GlobalConfig *SteampipeConfig
var defaultConfigFileName = "default.spc"
var defaultConfigSampleFileName = "default.spc.sample"

// LoadSteampipeConfig loads the HCL connection config and workspace options
func LoadSteampipeConfig(ctx context.Context, modLocation string, commandName string) (*SteampipeConfig, perror_helpers.ErrorAndWarnings) {
	utils.LogTime("steampipeconfig.LoadSteampipeConfig start")
	defer utils.LogTime("steampipeconfig.LoadSteampipeConfig end")

	log.Printf("[INFO] ensureDefaultConfigFile")

	if err := ensureDefaultConfigFile(pfilepaths.EnsureConfigDir()); err != nil {
		return nil, perror_helpers.NewErrorsAndWarning(
			sperr.WrapWithMessage(
				err,
				"could not create default config",
			),
		)
	}
	return loadSteampipeConfig(ctx, modLocation, commandName)
}

// LoadConnectionConfig loads the connection config but not the workspace options
// this is called by the fdw
func LoadConnectionConfig(ctx context.Context) (*SteampipeConfig, perror_helpers.ErrorAndWarnings) {
	return LoadSteampipeConfig(ctx, "", "")
}

func ensureDefaultConfigFile(configFolder string) error {
	// get the filepaths
	defaultConfigFile := filepath.Join(configFolder, defaultConfigFileName)
	defaultConfigSampleFile := filepath.Join(configFolder, defaultConfigSampleFileName)

	// check if sample and default files exist
	sampleExists := filehelpers.FileExists(defaultConfigSampleFile)
	defaultExists := filehelpers.FileExists(defaultConfigFile)

	var sampleContent []byte
	var sampleModTime, defaultModTime time.Time

	// if the sample file exists, load content and read mod time
	if sampleExists {
		sampleStat, err := os.Stat(defaultConfigSampleFile)
		if err != nil {
			return err
		}
		sampleContent, err = os.ReadFile(defaultConfigSampleFile)
		if err != nil {
			return err
		}
		sampleModTime = sampleStat.ModTime()
	}

	// if the default file exists read mod time
	if defaultExists {
		// get the file infos
		defaultStat, err := os.Stat(defaultConfigFile)
		if err != nil {
			return err
		}
		// get the file mod times
		defaultModTime = defaultStat.ModTime()
	}

	// check if the files are modified

	// has the user modified the default file?
	userModifiedDefault := defaultModTime.IsZero() ||
		defaultModTime.After(sampleModTime) && defaultModTime.Sub(sampleModTime) > 100*time.Millisecond

	// has the DefaultConnectionConfigContent been updated since the sample file was last writtne
	sampleModified := sampleModTime.IsZero() ||
		!bytes.Equal([]byte(constants.DefaultConnectionConfigContent), sampleContent)

	// case: if sample is modified - always write new sample file content
	if sampleModified {
		err := os.WriteFile(defaultConfigSampleFile, []byte(constants.DefaultConnectionConfigContent), 0755)
		if err != nil {
			return err
		}
	}

	// case: if sample is modified but default is not modified - write the new default file content
	if sampleModified && !userModifiedDefault {
		err := os.WriteFile(defaultConfigFile, []byte(constants.DefaultConnectionConfigContent), 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadSteampipeConfig(ctx context.Context, modLocation string, commandName string) (steampipeConfig *SteampipeConfig, errorsAndWarnings perror_helpers.ErrorAndWarnings) {
	utils.LogTime("steampipeconfig.loadSteampipeConfig start")
	defer utils.LogTime("steampipeconfig.loadSteampipeConfig end")

	errorsAndWarnings = perror_helpers.NewErrorsAndWarning(nil)
	defer func() {
		if r := recover(); r != nil {
			errorsAndWarnings = perror_helpers.NewErrorsAndWarning(helpers.ToError(r))
		}
	}()

	steampipeConfig = NewSteampipeConfig(commandName)

	// load plugin versions
	v, err := versionfile.LoadPluginVersionFile(ctx)
	if err != nil {
		return nil, perror_helpers.NewErrorsAndWarning(err)
	}

	// add any "local" plugins (i.e. plugins installed under the 'local' folder) into the version file
	ew := v.AddLocalPlugins(ctx)
	if ew.GetError() != nil {
		return nil, ew
	}
	steampipeConfig.PluginVersions = v.Plugins

	// load config from the installation folder -  load all spc files from config directory
	include := filehelpers.InclusionsFromExtensions(pconstants.ConnectionConfigExtensions)
	loadOptions := &loadConfigOptions{include: include}
	ew = loadConfig(ctx, pfilepaths.EnsureConfigDir(), steampipeConfig, loadOptions)
	if ew.GetError() != nil {
		return nil, ew
	}
	// merge the warning from this call
	errorsAndWarnings.AddWarning(ew.Warnings...)

	// now load config from the workspace folder, if provided
	// this has precedence and so will overwrite any config which has already been set
	// check workspace folder exists
	if modLocation != "" {
		if _, err := os.Stat(modLocation); os.IsNotExist(err) {
			return nil, perror_helpers.NewErrorsAndWarning(fmt.Errorf("mod location '%s' does not exist", modLocation))
		}

		// only include workspace.spc from workspace directory
		include = filehelpers.InclusionsFromFiles([]string{filepaths.WorkspaceConfigFileName})
		// update load options to ONLY allow terminal options
		loadOptions = &loadConfigOptions{include: include, allowedOptions: []string{options.TerminalBlock}}
		ew := loadConfig(ctx, modLocation, steampipeConfig, loadOptions)
		if ew.GetError() != nil {
			return nil, ew.WrapErrorWithMessage("failed to load workspace config")
		}

		// merge the warning from this call
		errorsAndWarnings.AddWarning(ew.Warnings...)
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

func loadConfig(ctx context.Context, configFolder string, steampipeConfig *SteampipeConfig, opts *loadConfigOptions) perror_helpers.ErrorAndWarnings {
	log.Printf("[INFO] loadConfig is loading connection config")
	// get all the config files in the directory
	configPaths, err := filehelpers.ListFilesWithContext(ctx, configFolder, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: opts.include,
	})

	if err != nil {
		log.Printf("[WARN] loadConfig: failed to get config file paths: %v\n", err)
		return perror_helpers.NewErrorsAndWarning(err)
	}
	if len(configPaths) == 0 {
		return perror_helpers.ErrorAndWarnings{}
	}

	fileData, diags := parse.LoadFileData(configPaths...)
	if diags.HasErrors() {
		log.Printf("[WARN] loadConfig: failed to load all config files: %v\n", err)
		return perror_helpers.DiagsToErrorsAndWarnings("Failed to load all config files", diags)
	}

	body, diags := parse.ParseHclFiles(fileData)
	if diags.HasErrors() {
		return perror_helpers.DiagsToErrorsAndWarnings("Failed to load all config files", diags)
	}

	// do a partial decode
	content, moreDiags := body.Content(parse.ConfigBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return perror_helpers.DiagsToErrorsAndWarnings("Failed to load config", diags)
	}

	// store block types which we have found in this folder - each is only allowed once
	// NOTE this is different to merging options with options already populated in the passed-in steampipe config
	// this is valid because the same block may be defined in the config folder and the workspace
	optionBlockMap := map[string]bool{}

	for _, block := range content.Blocks {
		switch block.Type {

		case modconfig.BlockTypePlugin:
			plugin, moreDiags := parse.DecodePlugin(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			// add plugin to steampipeConfig
			// NOTE: this errors if there is a plugin block with a duplicate label
			if err := steampipeConfig.addPlugin(plugin); err != nil {
				return perror_helpers.NewErrorsAndWarning(err)
			}

		case modconfig.BlockTypeConnection:
			connection, moreDiags := parse.DecodeConnection(block)
			diags = append(diags, moreDiags...)
			if moreDiags.HasErrors() {
				continue
			}
			if existingConnection, alreadyThere := steampipeConfig.Connections[connection.Name]; alreadyThere {
				err := getDuplicateConnectionError(existingConnection, connection)
				return perror_helpers.NewErrorsAndWarning(err)
			}
			if ok, errorMessage := db_common.IsSchemaNameValid(connection.Name); !ok {
				return perror_helpers.NewErrorsAndWarning(sperr.New("invalid connection name: '%s' in '%s'. %s ", connection.Name, block.TypeRange.Filename, errorMessage))
			}
			steampipeConfig.Connections[connection.Name] = connection

		case modconfig.BlockTypeOptions:
			// check this options type is permitted based on the options passed in
			if err := optionsBlockPermitted(block, optionBlockMap, opts); err != nil {
				return perror_helpers.NewErrorsAndWarning(err)
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
						Subject:  hclhelpers.BlockRangePointer(block),
					})
				}
			}
		}
	}

	if diags.HasErrors() {
		return perror_helpers.DiagsToErrorsAndWarnings("Failed to load config", diags)
	}

	res := perror_helpers.DiagsToErrorsAndWarnings("", diags)

	log.Printf("[INFO] loadConfig calling initializePlugins")

	// resolve the plugins for each connection and create default plugin config
	// for all plugins mentioned in connection config which have no explicit config
	steampipeConfig.initializePlugins()

	return res
}

func getDuplicateConnectionError(existingConnection, newConnection *modconfig.Connection) error {
	return sperr.New("duplicate connection name: '%s'\n\t(%s:%d)\n\t(%s:%d)",
		existingConnection.Name, existingConnection.DeclRange.Filename, existingConnection.DeclRange.Start.Line,
		newConnection.DeclRange.Filename, newConnection.DeclRange.Start.Line)
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
