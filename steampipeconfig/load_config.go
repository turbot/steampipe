package steampipeconfig

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/schema"
	"github.com/turbot/steampipe/steampipeconfig/options"
)

var Config *SteampipeConfig
var defaultConfigFileName = "default.spc"

// LoadSteampipeConfig :: load the HCL config and parse into the global Config variable
func LoadSteampipeConfig(workspacePath string) (*SteampipeConfig, error) {
	_ = ensureDefaultConfigFile(constants.ConfigDir())
	config, err := newSteampipeConfig(workspacePath)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func ensureDefaultConfigFile(configFolder string) error {
	defaultConfigFile := filepath.Join(configFolder, defaultConfigFileName)
	if _, err := os.Stat(defaultConfigFile); os.IsNotExist(err) {
		err = ioutil.WriteFile(defaultConfigFile, []byte(constants.DefaultSPCContent), 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func newSteampipeConfig(workspacePath string) (steampipeConfig *SteampipeConfig, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	steampipeConfig = &SteampipeConfig{
		Connections: make(map[string]*Connection),
	}

	// load config from the installation folder -  load all spc files from config directory
	include := filehelpers.InclusionsFromExtensions([]string{constants.ConfigExtension})
	loadOptions := &loadConfigOptions{include: include}
	if err := loadConfig(constants.ConfigDir(), steampipeConfig, loadOptions); err != nil {
		return nil, err
	}

	// At present, this function is used both by steampipe to load connection config AND options,
	// and by the fdw to load just the conneciton config
	// when the fdw calls itr, it will NOT pass a workspace path
	// TODO refactor this to enable loading connection config only for FDW

	// now load config from the workspace folder
	// this has precedence and so will overwrite any config which has already been set
	// check workspace folder exists
	if workspacePath != "" {
		if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("workspace folder '%s' does not exist", workspacePath)
		}

		// only include workspace.spc from workspace directory
		include = filehelpers.InclusionsFromFiles([]string{constants.WorkspaceConfigFileName})
		// update load options to ONLY allow terminal options
		loadOptions = &loadConfigOptions{include: include, allowedOptions: []string{options.TerminalBlock}}
		if err := loadConfig(workspacePath, steampipeConfig, loadOptions); err != nil {
			return nil, fmt.Errorf("failed to load workspace config: %v", err)
		}
	}

	// now set default options on all connections without options set
	steampipeConfig.setDefaultConnectionOptions()

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
	configPaths, err := filehelpers.ListFiles(configFolder, &filehelpers.ListFilesOptions{
		Options: filehelpers.FilesFlat,
		Include: opts.include,
	})

	if err != nil {
		log.Printf("[WARN] loadConfig: failed to get config file paths: %v\n", err)
		return err
	}
	if len(configPaths) == 0 {
		return nil
	}

	fileData, diags := loadFileData(configPaths)
	if diags.HasErrors() {
		log.Printf("[WARN] loadConfig: failed to load all config files: %v\n", err)
		return plugin.DiagsToError("Failed to load all config files", diags)
	}

	body, diags := parseHclFiles(fileData)
	if diags.HasErrors() {
		return plugin.DiagsToError("Failed to load all config files", diags)
	}

	// do a partial decode
	content, moreDiags := body.Content(configSchema)
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
		case "connection":
			connection, moreDiags := parseConnection(block, fileData)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				continue
			}
			_, alreadyThere := steampipeConfig.Connections[connection.Name]
			if alreadyThere {
				return fmt.Errorf("duplicate connection name: '%s' in '%s'", connection.Name, block.TypeRange.Filename)
			}
			if !schema.IsSchemaNameValid(connection.Name) {
				return fmt.Errorf("invalid connection name: '%s' in '%s'", connection.Name, block.TypeRange.Filename)
			}
			steampipeConfig.Connections[connection.Name] = connection

		case "options":
			// check this options type is permitted based on the options passed in
			if err := optionsBlockPermitted(block, optionBlockMap, opts); err != nil {
				return err
			}
			options, moreDiags := parseOptions(block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				continue
			}
			// set options on steampipe config
			// if options are already set, this will merge the new options over the top of the existing options
			// i.e. new options have precedence
			steampipeConfig.SetOptions(options)
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
