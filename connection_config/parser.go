package connection_config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/schema"
)

const configExtension = ".spc"

func Load() (*ConnectionConfig, error) {
	return loadConfig(constants.ConfigDir())
}

func loadConfig(configFolder string) (*ConnectionConfig, error) {
	var result = newConfig()

	// get all the config files in the directory
	configPaths, err := getConfigFilePaths(configFolder)
	if err != nil {
		return nil, err
	}
	if len(configPaths) == 0 {
		return &ConnectionConfig{}, nil
	}

	body, diags := parseConfigs(configPaths)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to load all config files: %s", diags.Error())
	}

	// do a partial decode
	content, _, moreDiags := body.PartialContent(configSchema)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
		return nil, diags
	}

	for _, block := range content.Blocks {
		switch block.Type {
		case "connection":
			connection, moreDiags := parseConnection(block)
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
		}
	}

	if diags.HasErrors() {
		return nil, fmt.Errorf(diags.Error())
	}
	return result, nil
}

func parseConfigs(configPaths []string) (hcl.Body, hcl.Diagnostics) {
	var parsedConfigFiles []*hcl.File

	var diags hcl.Diagnostics
	for _, configPath := range configPaths {
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("failed to read config file %s", configPath),
				Detail:   err.Error()})
			continue
		}

		file, moreDiags := hclsyntax.ParseConfig(data, configPath, hcl.Pos{Byte: 0, Line: 1, Column: 1})
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			continue
		}
		parsedConfigFiles = append(parsedConfigFiles, file)
	}

	return hcl.MergeFiles(parsedConfigFiles), diags
}

func parseConnection(block *hcl.Block) (*Connection, hcl.Diagnostics) {
	connectionBlock, rest, diags := block.Body.PartialContent(connectionSchema)
	if diags.HasErrors() {
		return nil, diags
	}

	connection := NewConnection()
	// convert this into a fully qualified name plugin name
	connection.Name = block.Labels[0]

	var plugin string
	diags = gohcl.DecodeExpression(connectionBlock.Attributes["plugin"].Expr, nil, &plugin)
	if diags.HasErrors() {
		return nil, diags
	}
	connection.Plugin = ociinstaller.NewSteampipeImageRef(plugin).DisplayImageRef()

	// now populate the dynamic connection config
	remainingAttributes, diags := rest.JustAttributes()

	for name, attribute := range remainingAttributes {
		if name != "connection" {
			var val string
			diags = gohcl.DecodeExpression(attribute.Expr, nil, &val)
			if diags.HasErrors() {
				return nil, diags
			}
			connection.Config[name] = val
		}
	}

	return connection, nil
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
