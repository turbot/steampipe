package modconfig

import "github.com/turbot/go-kit/helpers"

// NOTE: when adding a block type, be sure to update  QueryProviderBlocks/ReferenceBlocks/AllBlockTypes as needed
const (
	// require blocks
	BlockTypeSteampipe = "steampipe"
	BlockTypeMod       = "mod"
	BlockTypePlugin    = "plugin"
	// resource blocks
	BlockTypeQuery          = "query"
	BlockTypeControl        = "control"
	BlockTypeBenchmark      = "benchmark"
	BlockTypeDashboard      = "dashboard"
	BlockTypeContainer      = "container"
	BlockTypeChart          = "chart"
	BlockTypeCard           = "card"
	BlockTypeFlow           = "flow"
	BlockTypeGraph          = "graph"
	BlockTypeHierarchy      = "hierarchy"
	BlockTypeImage          = "image"
	BlockTypeInput          = "input"
	BlockTypeTable          = "table"
	BlockTypeText           = "text"
	BlockTypeLocals         = "locals"
	BlockTypeVariable       = "variable"
	BlockTypeParam          = "param"
	BlockTypeRequire        = "require"
	BlockTypeNode           = "node"
	BlockTypeEdge           = "edge"
	BlockTypeLegacyRequires = "requires"
	BlockTypeCategory       = "category"
	BlockTypeWith           = "with"

	// config blocks
	BlockTypeConnection       = "connection"
	BlockTypeOptions          = "options"
	BlockTypeWorkspaceProfile = "workspace"

	ResourceTypeSnapshot = "snapshot"
	AttributeArgs        = "args"
	AttributeQuery       = "query"
)

// QueryProviderBlocks is a list of block types which implement QueryProvider
var QueryProviderBlocks = []string{
	BlockTypeCard,
	BlockTypeChart,
	BlockTypeControl,
	BlockTypeEdge,
	BlockTypeFlow,
	BlockTypeGraph,
	BlockTypeHierarchy,
	BlockTypeImage,
	BlockTypeInput,
	BlockTypeQuery,
	BlockTypeNode,
	BlockTypeTable,
	BlockTypeWith,
}

// NodeAndEdgeProviderBlocks is a list of block types which implementnodeAndEdgeProvider
var NodeAndEdgeProviderBlocks = []string{
	BlockTypeHierarchy,
	BlockTypeFlow,
	BlockTypeGraph,
}

// ReferenceBlocks is a list of block types we store references for
var ReferenceBlocks = []string{
	BlockTypeMod,
	BlockTypeQuery,
	BlockTypeControl,
	BlockTypeBenchmark,
	BlockTypeDashboard,
	BlockTypeContainer,
	BlockTypeCard,
	BlockTypeChart,
	BlockTypeFlow,
	BlockTypeGraph,
	BlockTypeHierarchy,
	BlockTypeImage,
	BlockTypeInput,
	BlockTypeTable,
	BlockTypeText,
	BlockTypeParam,
	BlockTypeCategory,
	BlockTypeWith,
}

var ValidResourceItemTypes = []string{
	BlockTypeMod,
	BlockTypeQuery,
	BlockTypeControl,
	BlockTypeBenchmark,
	BlockTypeDashboard,
	BlockTypeContainer,
	BlockTypeChart,
	BlockTypeCard,
	BlockTypeFlow,
	BlockTypeGraph,
	BlockTypeHierarchy,
	BlockTypeImage,
	BlockTypeInput,
	BlockTypeTable,
	BlockTypeText,
	BlockTypeLocals,
	BlockTypeVariable,
	BlockTypeParam,
	BlockTypeRequire,
	BlockTypeNode,
	BlockTypeEdge,
	BlockTypeLegacyRequires,
	BlockTypeCategory,
	BlockTypeConnection,
	BlockTypeOptions,
	BlockTypeWorkspaceProfile,
	BlockTypeWith,
	// local is not an actual block name but is a resource type
	"local",
	// references
	"ref",
	// variables
	"var",
}

func IsValidResourceItemType(blockType string) bool {
	return helpers.StringSliceContains(ValidResourceItemTypes, blockType)
}
