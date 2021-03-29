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

// Load :: load the HCL config and parse into the global Config variable
func Load() (*SteampipeConfig, error) {
	_ = ensureDefaultConfigFile(constants.ConfigDir())
	config, err := loadConfig(constants.ConfigDir())
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

func loadConfig(configFolder string) (steampipeConfig *SteampipeConfig, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	steampipeConfig = newSteampipeConfig()

	// get all the config files in the directory
	configPaths, err := filehelpers.ListFiles(configFolder, &filehelpers.ListFilesOptions{
		// todo recursive?
		Options: filehelpers.FilesFlat,
		Include: []string{fmt.Sprintf("**/*%s", constants.ConfigExtension)},
	})

	if err != nil {
		log.Printf("[WARN] loadConfig: failed to get config file paths: %v\n", err)
		return nil, err
	}
	if len(configPaths) == 0 {
		return &SteampipeConfig{}, nil
	}

	fileData, diags := loadFileData(configPaths)
	if diags.HasErrors() {
		log.Printf("[WARN] loadConfig: failed to load all config files: %v\n", err)
		return nil, plugin.DiagsToError("Failed to load all config files", diags)
	}

	body, diags := parseHclFiles(fileData)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load all config files", diags)
	}

	// do a partial decode
	content, moreDiags := body.Content(configSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, plugin.DiagsToError("Failed to load config", diags)
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
				return nil, fmt.Errorf("duplicate connection name: '%s' in '%s'", connection.Name, block.TypeRange.Filename)
			}
			if !schema.IsSchemaNameValid(connection.Name) {
				return nil, fmt.Errorf("invalid connection name: '%s' in '%s'", connection.Name, block.TypeRange.Filename)
			}
			steampipeConfig.Connections[connection.Name] = connection

		case "options":
			options, moreDiags := parseOptions(block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				continue
			}
			steampipeConfig.SetOptions(options)
		}
	}

	if diags.HasErrors() {
		return nil, plugin.DiagsToError("Failed to load config", diags)
	}
	// now set default options on all connections without options set
	steampipeConfig.setDefaultConnectionOptions()

	return steampipeConfig, nil
}
