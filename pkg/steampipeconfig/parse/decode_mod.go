package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func decodeMod(block *hcl.Block, evalCtx *hcl.EvalContext, mod *modconfig.Mod) (*modconfig.Mod, *decodeResult) {
	res := newDecodeResult()

	//  retrieve the body content which complies with modBlockSchema
	//  - this will be used to handle attributes which need manual decoding
	// everything else will be implicitly decoded
	content, remain, diags := block.Body.PartialContent(ModBlockSchema)
	res.handleDecodeDiags(diags)

	// decode the body to populate all properties that can be automatically decoded
	diags = decodeHclBody(remain, evalCtx, mod, mod)
	res.handleDecodeDiags(diags)
	if !res.Success() {
		return mod, res
	}

	// now decode the require block
	require, requireRes := decodeRequireBlock(content, evalCtx)
	res.Merge(requireRes)
	if require != nil {
		mod.Require = require
	}

	return mod, res
}

func decodeRequireBlock(content *hcl.BodyContent, evalCtx *hcl.EvalContext) (*modconfig.Require, *decodeResult) {
	var res = newDecodeResult()

	block := getFirstBlockOfType(content.Blocks, modconfig.BlockTypeRequire)
	if block == nil {
		return nil, res
	}

	//  retrieve the body content which complies with modBlockSchema
	//  - this will be used to handle attributes which need manual decoding
	// everything else will be implicitly decoded
	content, _, diags := block.Body.PartialContent(RequireBlockSchema)
	res.handleDecodeDiags(diags)

	// decode the body into 'modContainer' to populate all properties that can be automatically decoded
	require := modconfig.NewRequire()

	// TODO KAI SET DECL RANGE

	// manually decode deprecated `steampipe` property
	// decode the body into 'modContainer' to populate all properties that can be automatically decoded
	diags = gohcl.DecodeBody(block.Body, evalCtx, require)
	res.handleDecodeDiags(diags)

	var deprecatedSteampipeVersionString string
	diags = decodeProperty(content, "steampipe", &deprecatedSteampipeVersionString, evalCtx)
	if deprecatedSteampipeVersionString != "" {
		require.Steampipe = &modconfig.SteampipeRequire{MinVersionString: deprecatedSteampipeVersionString}
		res.Diags = append(res.Diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "Property 'steampipe' is deprecated, use steampipe block instead",
		},
		)
	}
	// TODO CHECK mod version deprecations

	return require, res
}
