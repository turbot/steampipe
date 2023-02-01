package steampipeconfig

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/schema"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/options"
	"github.com/turbot/steampipe/pkg/steampipeconfig/parse"
	"github.com/turbot/steampipe/pkg/utils"
)

var GlobalConfig *SteampipeConfig
var defaultConfigFileName = "default.spc"

// LoadSteampipeConfig loads the HCL connection config and workspace options
func LoadSteampipeConfig(modLocation string, commandName string) (*SteampipeConfig, error) {
	utils.LogTime("steampipeconfig.LoadSteampipeConfig start")
	defer utils.LogTime("steampipeconfig.LoadSteampipeConfig end")

	_ = ensureDefaultConfigFile(filepaths.EnsureConfigDir())
	config, err := loadSteampipeConfig(modLocation, commandName)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// LoadConnectionConfig loads the connection config but not the workspace options
// this is called by the fdw
func LoadConnectionConfig() (*SteampipeConfig, error) {
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

func loadSteampipeConfig(modLocation string, commandName string) (steampipeConfig *SteampipeConfig, err error) {
	utils.LogTime("steampipeconfig.loadSteampipeConfig start")
	defer utils.LogTime("steampipeconfig.loadSteampipeConfig end")

	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	steampipeConfig = NewSteampipeConfig(commandName)

	// load config from the installation folder -  load all spc files from config directory
	include := filehelpers.InclusionsFromExtensions(constants.ConnectionConfigExtensions)
	loadOptions := &loadConfigOptions{include: include}
	if err := loadConfig(filepaths.EnsureConfigDir(), steampipeConfig, loadOptions); err != nil {
		return nil, err
	}

	// now load config from the workspace folder, if provided
	// this has precedence and so will overwrite any config which has already been set
	// check workspace folder exists
	if modLocation != "" {
		if _, err := os.Stat(modLocation); os.IsNotExist(err) {
			return nil, fmt.Errorf("mod location '%s' does not exist", modLocation)
		}

		// only include workspace.spc from workspace directory
		include = filehelpers.InclusionsFromFiles([]string{filepaths.WorkspaceConfigFileName})
		// update load options to ONLY allow terminal options
		loadOptions = &loadConfigOptions{include: include, allowedOptions: []string{options.TerminalBlock}}
		if err := loadConfig(modLocation, steampipeConfig, loadOptions); err != nil {
			return nil, fmt.Errorf("failed to load workspace config: %v", err)
		}
	}

	// now set default options on all connections without options set
	// this is needed as the connection config is also loaded by the FDW which has no access to viper
	steampipeConfig.setDefaultConnectionOptions()

	// now validate the config
	if err := steampipeConfig.Validate(); err != nil {
		return nil, err
	}
	return steampipeConfig, nil
}

// load config from the given folder and update steampipeConfig
// NOTE: this mutates steampipe config
type loadConfigOptions struct {
	include        []string
	allowedOptions []string
}

func loadConfig(configFolder string, steampipeConfig *SteampipeConfig, opts *loadConfigOptions) error {
	// get all the config files in the directory
	configPaths, err := filehelpers.ListFiles(configFolder, &filehelpers.ListOptions{
		Flags:   filehelpers.FilesFlat,
		Include: opts.include,
	})

	if err != nil {
		log.Printf("[WARN] loadConfig: failed to get config file paths: %v\n", err)
		return err
	}
	if len(configPaths) == 0 {
		return nil
	}

	fileData, diags := parse.LoadFileData(configPaths...)
	if diags.HasErrors() {
		log.Printf("[WARN] loadConfig: failed to load all config files: %v\n", err)
		return plugin.DiagsToError("Failed to load all config files", diags)
	}

	body, diags := parse.ParseHclFiles(fileData)
	if diags.HasErrors() {
		return plugin.DiagsToError("Failed to load all config files", diags)
	}

	// do a partial decode
	content, moreDiags := body.Content(parse.ConfigBlockSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return plugin.DiagsToError("Failed to load config", diags)
	}

	// store block types which we have found in this folder - each is only allowed once
	// NOTE this is different to merging options with options already populated in the passed-in steampipe config
	// this is valid because the same block may be defined in the config folder and the workspace
	optionBlockMap := map[string]bool{}

	for _, block := range content.Blocks {
		switch block.Type {
		case modconfig.BlockTypeConnection:
			connection, moreDiags := parse.DecodeConnection(block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				continue
			}

			srcRange := block.Body.(*hclsyntax.Body).SrcRange

			connection.Config = string(fileData[srcRange.Filename][srcRange.Start.Byte:srcRange.End.Byte])

			_, alreadyThere := steampipeConfig.Connections[connection.Name]
			if alreadyThere {
				return fmt.Errorf("duplicate connection name: '%s' in '%s'", connection.Name, block.TypeRange.Filename)
			}
			if ok, errorMessage := schema.IsSchemaNameValid(connection.Name); !ok {
				return fmt.Errorf("invalid connection name: '%s' in '%s'. %s ", connection.Name, block.TypeRange.Filename, errorMessage)
			}
			steampipeConfig.Connections[connection.Name] = connection

		case modconfig.BlockTypeOptions:
			// check this options type is permitted based on the options passed in
			if err := optionsBlockPermitted(block, optionBlockMap, opts); err != nil {
				return err
			}
			opts, moreDiags := parse.DecodeOptions(block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				continue
			}
			// set options on steampipe config
			// if options are already setπ, this will merge the new options over the top of the existing options
			// i.e. new options have precedence
			steampipeConfig.SetOptions(opts)
		}
	}

	if diags.HasErrors() {
		return plugin.DiagsToError("Failed to load config", diags)
	}
	return nil
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
