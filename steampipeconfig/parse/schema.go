package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

var ConfigBlockSchema = &hcl.BodySchema{
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

var ConnectionBlockSchema = &hcl.BodySchema{
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

// ModBlockSchema is the top level schema for all mod resources
var ModBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       string(modconfig.BlockTypeMod),
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeVariable,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeQuery,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeControl,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeBenchmark,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeDashboard,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeCard,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeChart,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeFlow,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeHierarchy,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeImage,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeInput,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeTable,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeText,
			LabelNames: []string{"name"},
		},
		{
			Type: modconfig.BlockTypeLocals,
		},
	},
}

// DashboardBlockSchema contains the Dashboard attributes which cannot be automatically decoded
var DashboardBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "args"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       modconfig.BlockTypeInput,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeParam,
			LabelNames: []string{"name"},
		},
		{
			Type: modconfig.BlockTypeContainer,
		},
		{
			Type: modconfig.BlockTypeCard,
		},
		{
			Type: modconfig.BlockTypeChart,
		},
		{
			Type: modconfig.BlockTypeBenchmark,
		},
		{
			Type: modconfig.BlockTypeControl,
		},
		{
			Type: modconfig.BlockTypeFlow,
		},
		{
			Type: modconfig.BlockTypeHierarchy,
		},
		{
			Type: modconfig.BlockTypeImage,
		},
		{
			Type: modconfig.BlockTypeTable,
		},
		{
			Type: modconfig.BlockTypeText,
		},
	},
}

// DashboardContainerBlockSchema contains the DashboardContainer attributes which cannot be automatically decoded
var DashboardContainerBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "args"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       modconfig.BlockTypeInput,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeParam,
			LabelNames: []string{"name"},
		},
		{
			Type: modconfig.BlockTypeContainer,
		},
		{
			Type: modconfig.BlockTypeCard,
		},
		{
			Type: modconfig.BlockTypeChart,
		},
		{
			Type: modconfig.BlockTypeBenchmark,
		},
		{
			Type: modconfig.BlockTypeControl,
		},
		{
			Type: modconfig.BlockTypeFlow,
		},
		{
			Type: modconfig.BlockTypeHierarchy,
		},
		{
			Type: modconfig.BlockTypeImage,
		},
		{
			Type: modconfig.BlockTypeTable,
		},
		{
			Type: modconfig.BlockTypeText,
		},
	},
}

var BenchmarkBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "children"},
		{Name: "description"},
		{Name: "documentation"},
		{Name: "tags"},
		{Name: "title"},
		// for report benchmark blocks
		{Name: "width"},
		{Name: "base"},
		{Name: "type"},
		{Name: "display"},
	},
}

// QueryProviderBlockSchema schema for all blocks satisfying QueryProvider interface
var QueryProviderBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "args"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "param",
			LabelNames: []string{"name"},
		},
	},
}

var ParamDefBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "description"},
		{Name: "default"},
	},
}

var VariableBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name: "description",
		},
		{
			Name: "default",
		},
		{
			Name: "type",
		},
		{
			Name: "sensitive",
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type: "validation",
		},
	},
}
