package parse

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/pipe-fittings/modconfig"
)

// cache resource schemas
var resourceSchemaCache = make(map[string]*hcl.BodySchema)

var ConfigBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       modconfig.BlockTypeConnection,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypePlugin,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeOptions,
			LabelNames: []string{"type"},
		},
		{
			Type:       modconfig.BlockTypeWorkspaceProfile,
			LabelNames: []string{"name"},
		},
	},
}
var PluginBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       modconfig.BlockTypeRateLimiter,
			LabelNames: []string{"name"},
		},
	},
}

var WorkspaceProfileBlockSchema = &hcl.BodySchema{

	Blocks: []hcl.BlockHeaderSchema{
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
		{
			Name: "import_schema",
		},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "options",
			LabelNames: []string{"type"},
		},
	},
}

// WorkspaceBlockSchema is the top level schema for all workspace resources
var WorkspaceBlockSchema = &hcl.BodySchema{
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
			Type:       modconfig.BlockTypeGraph,
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
			Type:       modconfig.BlockTypeNode,
			LabelNames: []string{"name"},
		},
		{
			Type:       modconfig.BlockTypeEdge,
			LabelNames: []string{"name"},
		},
		{
			Type: modconfig.BlockTypeLocals,
		},
		{
			Type:       modconfig.BlockTypeCategory,
			LabelNames: []string{"name"},
		},
	},
}

// DashboardBlockSchema is only used to validate the blocks of a Dashboard
var DashboardBlockSchema = &hcl.BodySchema{
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
			Type: modconfig.BlockTypeWith,
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
			Type: modconfig.BlockTypeGraph,
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

// DashboardContainerBlockSchema is only used to validate the blocks of a DashboardContainer
var DashboardContainerBlockSchema = &hcl.BodySchema{
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
			Type: modconfig.BlockTypeGraph,
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
// NOTE: these are just the blocks/attributes that are explicitly decoded
// other query provider properties are implicitly decoded using tags
var QueryProviderBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "args"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "param",
			LabelNames: []string{"name"},
		},
		{
			Type:       "with",
			LabelNames: []string{"name"},
		},
	},
}

// NodeAndEdgeProviderSchema is used to decode graph/hierarchy/flow
// (EXCEPT categories)
var NodeAndEdgeProviderSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "args"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "param",
			LabelNames: []string{"name"},
		},
		{
			Type:       "category",
			LabelNames: []string{"name"},
		},
		{
			Type:       "with",
			LabelNames: []string{"name"},
		},
		{
			Type: modconfig.BlockTypeNode,
		},
		{
			Type: modconfig.BlockTypeEdge,
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
