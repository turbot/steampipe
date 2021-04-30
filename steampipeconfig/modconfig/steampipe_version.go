package modconfig

import "github.com/hashicorp/hcl/v2"

type SteampipeVersion struct {
	Version   string `cty:"version"`
	DeclRange hcl.Range
}

func NewSteampipeVersion(block *hcl.Block) *SteampipeVersion {
	return &SteampipeVersion{
		Version:   block.Labels[0],
		DeclRange: block.DefRange,
	}
}
