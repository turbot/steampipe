package steampipeconfig

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	filehelpers "github.com/turbot/go-kit/files"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/schema"
)

var Config *SteampipeConfig
var defaultConfigFileName = "default.spc"

// LoadConfig :: load the HCL config and parse into the global Config variable
func LoadConfig(workspacePath string) (*SteampipeConfig, error) {
	_ = ensureDefaultConfigFile(constants.ConfigDir())
	config, err := newSteampipeConfig(workspacePath)
	if err != nil {
		return nil, err
	}
	Config = config
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

	// load config from the installation folder
	// load all spc files from config directory
	include := filehelpers.InclusionsFromExtensions([]string{constants.ConfigExtension})
	if err := loadConfig(constants.ConfigDir(), steampipeConfig, include); err != nil {
		return nil, err
	}

	// now load config from the workspace folder
	//- this has precedence and so will overwrite any config which has already been set
	// only include workspace.spc from workspace directory
	include = filehelpers.InclusionsFromFiles([]string{constants.WorkspaceConfigFileName})
	if err := loadConfig(workspacePath, steampipeConfig, include); err != nil {
		return nil, err
	}

	// now set default options on all connections without options set
	steampipeConfig.setDefaultConnectionOptions()

	return steampipeConfig, nil
}

// load config from the given folder and update steampipeConfig
// NOTE: this mutates steampipe config

func loadConfig(configFolder string, steampipeConfig *SteampipeConfig, include []string) error {
	// get all the config files in the directory
	configPaths, err := filehelpers.ListFiles(configFolder, &filehelpers.ListFilesOptions{
		Options: filehelpers.FilesRecursive,
		Include: include,
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
