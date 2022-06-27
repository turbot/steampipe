package parse

import "github.com/hashicorp/hcl/v2"

func getFirstBlockOfType(blocks hcl.Blocks, blockType string) *hcl.Block {
	for _, block := range blocks {
		if block.Type == blockType {
			return block
		}
	}
	return nil
}
