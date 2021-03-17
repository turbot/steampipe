package steampipeconfig

import (
	"fmt"
	"log"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/schema"
)

var Config *SteampipeConfig

// Load :: load the HCL config and parse into the global Config variable
func Load() error {
	config, err := loadConfig(constants.ConfigDir())
	if err != nil {
		return err
	}
	Config = config
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
	configPaths, err := getConfigFilePaths(configFolder)
	if err != nil {
		log.Printf("[WARN] loadConfig: failed to get config file paths: %v\n", err)
		return nil, err
	}
	if len(configPaths) == 0 {
		log.Println("[DEBUG] loadConfig: 0 config file paths returned")
		return &SteampipeConfig{}, nil
	}

	fileData, diags := loadFileData(configPaths)
	if diags.HasErrors() {
		log.Printf("[WARN] loadConfig: failed to load all config files: %v\n", err)
		return nil, plugin.DiagsToError("failed to load all config files", diags)
	}

	body, diags := parseConfigs(fileData)
	if diags.HasErrors() {
		return nil, plugin.DiagsToError("failed to load all config files", diags)
	}

	// do a partial decode
	content, _, moreDiags := body.PartialContent(configSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, plugin.DiagsToError("failed to decode config", diags)
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
		return nil, plugin.DiagsToError("failed to load config", diags)
	}
	// now set default options on all connections without options set
	steampipeConfig.setDefaultConnectionOptions()

	return steampipeConfig, nil
}
