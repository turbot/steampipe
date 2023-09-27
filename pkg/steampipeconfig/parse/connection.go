package parse

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig/hclhelpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/exp/maps"
)

func DecodeConnection(block *hcl.Block) (*modconfig.Connection, hcl.Diagnostics) {
	connectionContent, rest, diags := block.Body.PartialContent(ConnectionBlockSchema)
	if diags.HasErrors() {
		return nil, diags
	}

	connection := modconfig.NewConnection(block)

	// decode the plugin property
	// NOTE: this mutates connection to set PluginAlias and possible PluginInstance
	diags = decodeConnectionPluginProperty(connectionContent, connection)
	if diags.HasErrors() {
		return nil, diags
	}

	if connectionContent.Attributes["type"] != nil {
		var connectionType string
		diags = gohcl.DecodeExpression(connectionContent.Attributes["type"].Expr, nil, &connectionType)
		if diags.HasErrors() {
			return nil, diags
		}
		connection.Type = connectionType
	}
	if connectionContent.Attributes["import_schema"] != nil {
		var importSchema string
		diags = gohcl.DecodeExpression(connectionContent.Attributes["import_schema"].Expr, nil, &importSchema)
		if diags.HasErrors() {
			return nil, diags
		}
		connection.ImportSchema = importSchema
	}
	if connectionContent.Attributes["connections"] != nil {
		var connections []string
		diags = gohcl.DecodeExpression(connectionContent.Attributes["connections"].Expr, nil, &connections)
		if diags.HasErrors() {
			return nil, diags
		}
		connection.ConnectionNames = connections
	}

	// if this is hcl config, check for nested blocks
	if body, ok := rest.(*hclsyntax.Body); ok {
		for _, connectionBlock := range body.Blocks {
			switch connectionBlock.Type {
			case "options":
				// if we already found settings, fail
				opts, moreDiags := DecodeOptions(connectionBlock.AsHCLBlock())
				if moreDiags.HasErrors() {
					diags = append(diags, moreDiags...)
					break
				}
				moreDiags = connection.SetOptions(opts, connectionBlock.AsHCLBlock())
				if moreDiags.HasErrors() {
					diags = append(diags, moreDiags...)
				}

				// TODO: remove in 0.22 [https://github.com/turbot/steampipe/issues/3251]
				if connection.Options != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagWarning,
						Summary:  fmt.Sprintf("%s in %s have been deprecated and will be removed in subsequent versions of steampipe", constants.Bold("'connection' options"), constants.Bold("'connection' blocks")),
						Subject:  hclhelpers.BlockRangePointer(connectionBlock.AsHCLBlock()),
					})
				}

			default:
				// raise error for any other blocks
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("connections do not support '%s' blocks", connectionBlock.Type),
					Subject:  hclhelpers.BlockRangePointer(connectionBlock.AsHCLBlock()),
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

func decodeConnectionPluginProperty(connectionContent *hcl.BodyContent, connection *modconfig.Connection) hcl.Diagnostics {
	var pluginName string
	evalCtx := &hcl.EvalContext{Variables: make(map[string]cty.Value)}

	diags := gohcl.DecodeExpression(connectionContent.Attributes["plugin"].Expr, evalCtx, &pluginName)
	res := newDecodeResult()
	res.handleDecodeDiags(diags)
	if res.Diags.HasErrors() {
		return res.Diags
	}
	if len(res.Depends) > 0 {
		log.Printf("[INFO] decodeConnectionPluginProperty plugin property is HCL reference")
		// if this is a plugin reference, extract the plugin instance
		pluginInstance, ok := getPluginInstanceFromDependency(maps.Values(res.Depends))
		if !ok {
			log.Printf("[INFO] failed to resolve plugin property")
			// return the original diagnostics
			return diags
		}

		// so we have resolved a reference to a plugin config
		// we will validate that this block exists later in initializePlugins
		// set PluginInstance ONLY
		// (the PluginInstance property being set means that we will raise the correct error if we fail to resolve the plugin block)
		connection.PluginInstance = &pluginInstance
		return nil
	}

	// NOTE: plugin property is set in initializePlugins
	connection.PluginAlias = pluginName

	return nil
}

func getPluginInstanceFromDependency(dependencies []*modconfig.ResourceDependency) (string, bool) {
	if len(dependencies) != 1 {
		return "", false
	}
	if len(dependencies[0].Traversals) != 1 {
		return "", false
	}
	traversalString := hclhelpers.TraversalAsString(dependencies[0].Traversals[0])
	split := strings.Split(traversalString, ".")
	if len(split) != 2 || split[0] != "plugin" {
		return "", false
	}
	return split[1], true
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
	var sortedKeys = helpers.SortedMapKeys(attrExpressionMap)
	for _, name := range sortedKeys {
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
