package parse

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/turbot/steampipe/ociinstaller"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

func DecodeConnection(block *hcl.Block, fileData map[string][]byte) (*modconfig.Connection, hcl.Diagnostics) {
	connectionContent, rest, diags := block.Body.PartialContent(ConnectionBlockSchema)
	if diags.HasErrors() {
		return nil, diags
	}

	// get connection name
	connection := modconfig.NewConnection(block)

	var pluginName string
	diags = gohcl.DecodeExpression(connectionContent.Attributes["plugin"].Expr, nil, &pluginName)
	if diags.HasErrors() {
		return nil, diags
	}
	connection.Plugin = ociinstaller.NewSteampipeImageRef(pluginName).DisplayImageRef()
	connection.PluginShortName = pluginName

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
		connection.ConnectionNames = connections
	}

	// check for nested options
	for _, connectionBlock := range connectionContent.Blocks {
		switch connectionBlock.Type {
		case "options":
			// if we already found settings, fail
			opts, moreDiags := DecodeOptions(connectionBlock)
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
	// convert the remaining config to a hcl string to pass to the plugin
	config, moreDiags := BodyToHclString(rest, connectionContent)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
	} else {
		connection.Config = config
	}

	return connection, diags
}

func BodyToHclString(body hcl.Body, connectionContent *hcl.BodyContent) (string, hcl.Diagnostics) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	attrs, diags := body.JustAttributes()
	if diags.HasErrors() {
		return "", diags
	}
	for name, attr := range attrs {
		if _, ok := connectionContent.Attributes[name]; !ok {
			val, moreDiags := attr.Expr.Value(nil)
			diags = append(diags, moreDiags...)
			rootBody.SetAttributeValue(name, val) // this is overwritten later
		}
	}

	return string(f.Bytes()), diags
}
