package connection_config

import "github.com/hashicorp/hcl/v2"

var configSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "connection",
			LabelNames: []string{"name"},
		},
		{
			Type: "settings",
		},
	},
}

var connectionSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "plugin",
			Required: true,
		},
		{
			Name:     "cache",
			Required: false,
		},
		{
			Name:     "cache_ttl",
			Required: false,
		},
	},
}
