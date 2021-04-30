package modconfig

import "github.com/hashicorp/hcl/v2"

// Requires :: struct mod dependencies
type Requires struct {
	Steampipe *SteampipeVersion
	Plugins   []*PluginVersion
	Mods      []*ModVersion
	DeclRange hcl.Range
}

type RequiresConfig struct {
	Steampipe hcl.Block   `cty:"steampipe,block"`
	Plugins   []hcl.Block `cty:"plugin,block"`
	Mods      []hcl.Block `cty:"mod,block"`
	DeclRange hcl.Range
}

func (r *Requires) Schema() *hcl.BodySchema {
	return &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       BlockTypeModVersion,
				LabelNames: []string{"name", "version"},
			},
			{
				Type:       BlockTypePluginVersion,
				LabelNames: []string{"name", "version"},
			},
			{
				Type: BlockTypeSteampipeVersion,
				//LabelNames: []string{"version"},
			},
		},
	}
}
