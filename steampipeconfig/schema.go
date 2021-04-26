package steampipeconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

var configSchema = &hcl.BodySchema{
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

var connectionSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "plugin",
			Required: true,
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "options",
			LabelNames: []string{"type"},
		},
	},
}

var modFileSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "variable",
			LabelNames: []string{"name"},
		},
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
			Type:       string(modconfig.BlockTypeControlGroup),
			LabelNames: []string{"name"},
		},
	},
}

var modSchema = &hcl.BodySchema{
	// NOTE: these must be consistent with the properties of the Mod struct steampipeconfig/modconfig/mod.go
	Attributes: []hcl.AttributeSchema{
		{
			Name: "color",
		},
		{
			Name: "description",
		},
		{
			Name: "documentation",
		},
		{
			Name: "description",
		},
		{
			Name: "icon",
		},
		{
			Name: "description",
		},
		{
			Name: "labels",
		},
		{
			Name: "title",
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "requires",
		},
		{
			Type: "opengraph",
		},
	},
}
