package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

var ConfigSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "connection",
			LabelNames: []string{"name"},
		},
		{
			Type:       "options",
			LabelNames: []string{"type"},
		},
	},
}

var ConnectionSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "plugin",
			Required: true,
		},
		{
			Name: "type",
		},
		{
			Name: "connections",
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "options",
			LabelNames: []string{"type"},
		},
	},
}

// ModFileSchema :: top level schema for all mod resources
var ModFileSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       string(modconfig.BlockTypeMod),
			LabelNames: []string{"name"},
		},
		{
			Type:       string(modconfig.BlockTypeQuery),
			LabelNames: []string{"name"},
		},
		{
			Type:       string(modconfig.BlockTypeControl),
			LabelNames: []string{"name"},
		},
		{
			Type:       string(modconfig.BlockTypeBenchmark),
			LabelNames: []string{"name"},
		},
		{
			Type: string(modconfig.BlockTypeLocals),
		},
	},
}
