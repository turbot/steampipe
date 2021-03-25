package steampipeconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/steampipeconfig/options"
)

// built a map of filepath to file data
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

func parseHclFiles(fileData map[string][]byte) (hcl.Body, hcl.Diagnostics) {
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
	connectionContent, rest, diags := block.Body.PartialContent(connectionSchema)
	if diags.HasErrors() {
		return nil, diags
	}

	// get connection name
	connection := &Connection{Name: block.Labels[0]}

	var pluginName string
	diags = gohcl.DecodeExpression(connectionContent.Attributes["plugin"].Expr, nil, &pluginName)
	if diags.HasErrors() {
		return nil, diags
	}
	connection.Plugin = ociinstaller.NewSteampipeImageRef(pluginName).DisplayImageRef()

	// check for nested options
	for _, connectionBlock := range connectionContent.Blocks {
		switch connectionBlock.Type {
		case "options":
			// if we already found settings, fail
			opts, moreDiags := parseOptions(connectionBlock)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				break
			}
			connection.setOptions(opts, connectionBlock)

		default:
			// this can probably never happen
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("invalid block type '%s' - only 'options' blocks are supported for Connections", connectionBlock.Type),
				Subject:  &connectionBlock.DefRange,
			})
		}
	}
	// now build a string containing the hcl for all other connection config properties
	restBody := rest.(*hclsyntax.Body)
	var configProperties []string
	for name, a := range restBody.Attributes {
		// if this attribute does not appear in connectionContent, load the hcl string
		if _, ok := connectionContent.Attributes[name]; !ok {
			configProperties = append(configProperties, string(a.SrcRange.SliceBytes(fileData[a.SrcRange.Filename])))
		}
	}
	sort.Strings(configProperties)
	connection.Config = strings.Join(configProperties, "\n")

	return connection, diags
}

func parseOptions(block *hcl.Block) (options.Options, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	var dest options.Options
	switch block.Labels[0] {
	case options.ConnectionBlock:
		dest = &options.Connection{}
	case options.DatabaseBlock:
		dest = &options.Database{}
	case options.TerminalBlock:
		dest = &options.Terminal{}
	case options.GeneralBlock:
		dest = &options.General{}
	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Invalid options type '%s'", block.Labels[0]),
			Subject:  &block.DefRange,
		})
		return nil, diags
	}

	diags = gohcl.DecodeBody(block.Body, nil, dest)
	if diags.HasErrors() {
		return nil, diags
	}

	return dest, nil
}

// return a list of file paths of all files in given folder with given extension
func getFilePaths(configFolder, extension string) ([]string, error) {
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
		if filepath.Ext(entry.Name()) == extension {
			matches = append(matches, filepath.Join(configFolder, entry.Name()))
		}
	}
	return matches, nil
}
