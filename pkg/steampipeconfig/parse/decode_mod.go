package parse

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

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
	diags = gohcl.DecodeBody(remain, evalCtx, mod)
	// handle any resulting diags, which may specify dependencies
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
	content, remain, diags := block.Body.PartialContent(RequireBlockSchema)
	res.handleDecodeDiags(diags)

	// decode the body into 'modContainer' to populate all properties that can be automatically decoded
	require := modconfig.NewRequire()
	diags = gohcl.DecodeBody(remain, evalCtx, require)

	// handle any resulting diags, which may specify dependencies
	res.handleDecodeDiags(diags)

	modversionConstraints, modRes := decodeRequireModVersionConstraintBlocks(content, evalCtx)
	res.Merge(modRes)
	if modversionConstraints != nil {
		require.Mods = modversionConstraints
	}
	return require, res

}

func decodeRequireModVersionConstraintBlocks(content *hcl.BodyContent, evalCtx *hcl.EvalContext) ([]*modconfig.ModVersionConstraint, *decodeResult) {
	var res = newDecodeResult()
	var constraints []*modconfig.ModVersionConstraint

	for _, block := range content.Blocks {
		// we only expect mod blocks
		if block.Type != modconfig.BlockTypeMod {
			continue
		}

		//  retrieve the body content which complies with modBlockSchema
		//  - this will be used to handle attributes which need manual decoding
		// everything else will be implicitly decoded
		requireModContent, remain, diags := block.Body.PartialContent(RequireModBlockSchema)
		res.handleDecodeDiags(diags)

		// decode the body into 'modContainer' to populate all properties that can be automatically decoded
		constraint, _ := modconfig.NewModVersionConstraint(block.Labels[0])

		diags = gohcl.DecodeBody(remain, evalCtx, constraint)
		// handle any resulting diags, which may specify dependencies
		res.handleDecodeDiags(diags)

		args, modRes := decodeRequireModArgs(requireModContent, evalCtx)
		res.Merge(modRes)
		if args != nil {
			constraint.Args = args
		}
		constraints = append(constraints, constraint)
	}
	return constraints, res
}

func decodeRequireModArgs(content *hcl.BodyContent, evalCtx *hcl.EvalContext) (map[string]cty.Value, *decodeResult) {
	var res = newDecodeResult()

	attr, ok := content.Attributes["args"]
	if !ok {
		return nil, res
	}

	// try to evaluate expression
	val, diags := attr.Expr.Value(evalCtx)
	// handle any resulting diags, which may specify dependencies
	res.handleDecodeDiags(diags)
	if diags.HasErrors() {
		return nil, res
	}
	argMap, _ := ctyObjectToCtyArgMap(val)
	return argMap, res

}

func ctyObjectToCtyArgMap(val cty.Value) (map[string]cty.Value, error) {
	res := make(map[string]cty.Value)
	it := val.ElementIterator()
	for it.Next() {
		k, v := it.Element()

		// decode key
		var key string
		if err := gocty.FromCtyValue(k, &key); err != nil {
			return nil, err
		}

		if v.IsKnown() {
			res[key] = v
		}
	}
	return res, nil
}
