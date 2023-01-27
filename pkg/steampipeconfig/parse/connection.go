package parse

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/turbot/steampipe/pkg/ociinstaller"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func DecodeConnection(block *hcl.Block) (*modconfig.Connection, hcl.Diagnostics) {
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

	if strings.HasPrefix(pluginName, "local/") {
		connection.Plugin = pluginName
	} else {
		connection.Plugin = ociinstaller.NewSteampipeImageRef(pluginName).DisplayImageRef()
	}
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

	// NOTE: 'option' blocks are included in the ConnectionBlockSchema so we can parse out of connectionContent
	// however 'table' blocks are not in the schema - this is because the label is optional,
	// something not supported when using a block schema - so we decode those from the 'rest'
	for _, connectionBlock := range connectionContent.Blocks {
		if connectionBlock.Type != modconfig.BlockTypeOptions {
			// not expected - ConnectionBlockSchema only defines options
			panic(fmt.Sprintf("unexpected block type %s in decoded connection config", connectionBlock.Type))
		}

		opts, moreDiags := DecodeOptions(connectionBlock)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
			break
		}
		moreDiags = connection.SetOptions(opts, connectionBlock)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		}
	}
	// now look for table blocks in `rest`
	// NOTE: only supported for hcl config, NOT yml
	if restBody, ok := rest.(*hclsyntax.Body); ok {
		for _, connectionBlock := range restBody.Blocks {
			switch connectionBlock.Type {
			case modconfig.BlockTypeTable:
				// table block is only valid for aggregator connection
				if connection.Type != modconfig.ConnectionTypeAggregator {
					diags.Append(&hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "only aggregator connections can define 'table' blocks",
						Subject:  &block.DefRange})
					break
				}
				tableAggregationSpec, moreDiags := decodeTableAggregationSpec(connectionBlock.AsHCLBlock())
				if moreDiags.HasErrors() {
					diags = append(diags, moreDiags...)
					break
				}
				connection.TableAggregationSpecs = append(connection.TableAggregationSpecs, tableAggregationSpec)
			case modconfig.BlockTypeOptions:
				// ignore
			default:
				subject := connectionBlock.DefRange()
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("invalid block type '%s' - only 'options' blocks are supported for Connections", connectionBlock.Type),
					Subject:  &subject,
				})
			}
		}
	}
	// convert the remaining config to a hcl string to pass to the plugin
	config, moreDiags := pluginConnectionConfigToHclString(rest, connectionContent)
	if moreDiags.HasErrors() {
		diags = append(diags, moreDiags...)
	} else {
		connection.Config = config
	}

	return connection, diags

}

// DecodeOptions decodes an options block
func decodeTableAggregationSpec(block *hcl.Block) (*modconfig.TableAggregationSpec, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	res := &modconfig.TableAggregationSpec{}

	diags = gohcl.DecodeBody(block.Body, nil, res)
	if diags.HasErrors() {
		return nil, diags
	}

	if len(block.Labels) > 0 {
		// it is NOT valid for a table blockto have both a label and a match attribute
		if res.Match != "" {
			diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "table blocks cannot have both a label and a 'match' attribute",
				Subject:  &block.DefRange})
		}
		res.Match = block.Labels[0]
	}

	return res, nil
}

// build a hcl string with all attributes in the conneciton config which are NOT specified in the coneciton block schema
// this is passed to the plugin who will validate and parse it
func pluginConnectionConfigToHclString(body hcl.Body, connectionContent *hcl.BodyContent) (string, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	// this is a bit messy
	// we want to extract the attributes which are NOT in the connection block schema
	// the body passed in here is the 'rest' result returned from a partial decode, meaning all attributes and blocks
	// in the schema are marked as 'hidden'

	// body.JustAttributes() returns all attributes which are not hidden (i.e. all attributes NOT in the schema)
	//
	// however when calling JustAttributes for a hcl body, it will fail if there are any blocks
	// therefore this code will fail for hcl connection config which has any child blocks (e.g  connection options)
	//
	// it does work however for a json body as this implementation treats blocks as attributes,
	// so the options block is treated as a hidden attribute and excluded
	// we therefore need to treaty hcl and json body separately

	// store map of attribute expressions
	attrExpressionMap := make(map[string]hcl.Expression)

	if hclBody, ok := body.(*hclsyntax.Body); ok {
		// if we can cast to a hcl body, read all the attributes and manually exclude those which are in the schema
		for name, attr := range hclBody.Attributes {
			// exclude attributes we have already handled
			if _, ok := connectionContent.Attributes[name]; !ok {
				attrExpressionMap[name] = attr.Expr
			}
		}
	} else {
		// so the body was not hcl - we assume it is json
		// try to call JustAttributes
		attrs, diags := body.JustAttributes()
		if diags.HasErrors() {
			return "", diags
		}
		// the attributes returned will only be the ones not in the schema, i.e. we do not need to filter them ourselves
		for name, attr := range attrs {
			attrExpressionMap[name] = attr.Expr
		}
	}

	// build ordered list attributes
	// when we have generics we can add a GetOrderedMapKeys function
	var keys = make([]string, len(attrExpressionMap))
	i := 0
	for k := range attrExpressionMap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	for _, name := range keys {
		expr := attrExpressionMap[name]
		val, moreDiags := expr.Value(nil)
		if moreDiags.HasErrors() {
			diags = append(diags, moreDiags...)
		} else {
			rootBody.SetAttributeValue(name, val)
		}
	}

	return string(f.Bytes()), diags
}
