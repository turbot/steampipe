package connection_config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/schema"
)

const configExtension = ".spc"

func Load() (*SteampipeConfig, error) {
	return loadConfig(constants.ConfigDir())
}

func loadConfig(configFolder string) (result *SteampipeConfig, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
	}()

	result = newSteampipeConfig()

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
		log.Printf("[WARN] loadConfig: failed to get config file paths: %v\n", err)

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
			_, alreadyThere := result.Connections[connection.Name]
			if alreadyThere {
				return nil, fmt.Errorf("duplicate connection name: '%s' in '%s'", connection.Name, block.TypeRange.Filename)
			}
			if !schema.IsSchemaNameValid(connection.Name) {
				return nil, fmt.Errorf("invalid connection name: '%s' in '%s'", connection.Name, block.TypeRange.Filename)
			}
			result.Connections[connection.Name] = connection

		case "options":
			// if we already found settings, fail
			options, moreDiags := parseOptions(block)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				continue
			}
			result.SetOptions(options)
		}
	}

	if diags.HasErrors() {
		return nil, plugin.DiagsToError("failed to load config", diags)
	}

	return result, nil
}

func loadFileData(configPaths []string) (map[string][]byte, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var fileData = map[string][]byte{}

	for _, configPath := range configPaths {
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("failed to read config file %s", configPath),
				Detail:   err.Error()})
			continue
		}
		fileData[configPath] = data
	}
	return fileData, diags
}

func parseConfigs(fileData map[string][]byte) (hcl.Body, hcl.Diagnostics) {
	var parsedConfigFiles []*hcl.File
	var diags hcl.Diagnostics
	parser := hclparse.NewParser()
	for configPath, data := range fileData {
		file, moreDiags := parser.ParseHCL(data, configPath)

		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			continue
		}
		parsedConfigFiles = append(parsedConfigFiles, file)
	}

	return hcl.MergeFiles(parsedConfigFiles), diags
}

func parseConnection(block *hcl.Block, fileData map[string][]byte) (*Connection, hcl.Diagnostics) {
	connectionBlock, rest, diags := block.Body.PartialContent(connectionSchema)
	if diags.HasErrors() {
		return nil, diags
	}

	// get connection name
	connection := &Connection{Name: block.Labels[0]}

	var pluginName string
	diags = gohcl.DecodeExpression(connectionBlock.Attributes["plugin"].Expr, nil, &pluginName)
	if diags.HasErrors() {
		return nil, diags
	}
	connection.Plugin = ociinstaller.NewSteampipeImageRef(pluginName).DisplayImageRef()

	//// now evaluate other optional connection config properties
	//diags = decodeAttribute(connectionBlock, "cache", &connection.Cache)
	//if diags.HasErrors() {
	//	return nil, diags
	//}
	//diags = decodeAttribute(connectionBlock, "cache_ttl", &connection.CacheTTL)
	//if diags.HasErrors() {
	//	return nil, diags
	//}

	// now build a string containing the hcl for all other connection config properties
	restBody := rest.(*hclsyntax.Body)
	var configProperties []string
	for name, a := range restBody.Attributes {
		// if this attribute does not appear in connectionBlock, load the hcl string
		if _, ok := connectionBlock.Attributes[name]; !ok {
			configProperties = append(configProperties, string(a.SrcRange.SliceBytes(fileData[a.SrcRange.Filename])))
		}
	}
	connection.Config = strings.Join(configProperties, "\n")

	return connection, nil
}

func decodeAttribute(connectionBlock *hcl.BodyContent, property string, dest interface{}) hcl.Diagnostics {
	var diags hcl.Diagnostics
	if attr := connectionBlock.Attributes[property]; attr != nil {
		diags = gohcl.DecodeExpression(attr.Expr, nil, dest)
	}
	return diags
}

func parseOptions(block *hcl.Block) (Options, hcl.Diagnostics) {
	var dest Options
	switch block.Labels[0] {
	case HclOptionsConnection:
		dest = &ConnectionOptions{}
	case HclOptionsDatabase:
		dest = &DatabaseOptions{}
	case HclOptionsConsole:
		dest = &ConsoleOptions{}
	case HclOptionsGeneral:
		dest = &GeneralOptions{}
	}

	diags := gohcl.DecodeBody(block.Body, nil, dest)
	if diags.HasErrors() {
		return nil, diags
	}

	return dest, nil
}

func getConfigFilePaths(configFolder string) ([]string, error) {
	// check folder exists - if not just return empty config
	if _, err := os.Stat(configFolder); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := ioutil.ReadDir(configFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to read config folder %s: %v", configFolder, err)
	}

	matches := []string{}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == configExtension {
			matches = append(matches, filepath.Join(configFolder, entry.Name()))
		}
	}
	return matches, nil
}
