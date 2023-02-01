package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func BlockRange(block *hcl.Block) hcl.Range {
	res := block.DefRange
	if b, ok := block.Body.(*hclsyntax.Body); ok {
		res.End = b.SrcRange.End
	}
	return res
}
