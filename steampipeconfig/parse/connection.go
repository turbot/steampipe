package parse

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

func ParseConnection(block *hcl.Block, fileData map[string][]byte) (*modconfig.Connection, hcl.Diagnostics) {
	connectionContent, rest, diags := block.Body.PartialContent(ConnectionSchema)
	if diags.HasErrors() {
		return nil, diags
	}

	// get connection name
	connection := &modconfig.Connection{Name: block.Labels[0]}

	var pluginName string
	diags = gohcl.DecodeExpression(connectionContent.Attributes["plugin"].Expr, nil, &pluginName)
	if diags.HasErrors() {
		return nil, diags
	}
	connection.Plugin = ociinstaller.NewSteampipeImageRef(pluginName).DisplayImageRef()

	if connectionContent.Attributes["type"] != nil {
		var connectionType string
		diags = gohcl.DecodeExpression(connectionContent.Attributes["type"].Expr, nil, &connectionType)
		if diags.HasErrors() {
			return nil, diags
		}
		connection.Type = connectionType
	}
	if connectionContent.Attributes["connections"] != nil {
		var connections []string
		diags = gohcl.DecodeExpression(connectionContent.Attributes["connections"].Expr, nil, &connections)
		if diags.HasErrors() {
			return nil, diags
		}
		connection.Connections = connections
	}

	// check for nested options
	for _, connectionBlock := range connectionContent.Blocks {
		switch connectionBlock.Type {
		case "options":
			// if we already found settings, fail
			opts, moreDiags := ParseOptions(connectionBlock)
			if moreDiags.HasErrors() {
				diags = append(diags, moreDiags...)
				break
			}
			connection.SetOptions(opts, connectionBlock)

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
