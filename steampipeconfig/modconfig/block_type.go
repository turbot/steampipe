package modconfig

const (
	BlockTypeMod       = "mod"
	BlockTypeQuery     = "query"
	BlockTypeControl   = "control"
	BlockTypeBenchmark = "benchmark"
	BlockTypeDashboard = "dashboard"
	BlockTypeContainer = "container"
	BlockTypeChart     = "chart"
	BlockTypeCard      = "card"
	BlockTypeHierarchy = "hierarchy"
	BlockTypeImage     = "image"
	BlockTypeInput     = "input"
	BlockTypeTable     = "table"
	BlockTypeText      = "text"
	BlockTypeLocals    = "locals"
	BlockTypeVariable  = "variable"
	BlockTypeParam     = "param"
)

// QueryProviderBlocks is a list of block types which implement QueryProvider
var QueryProviderBlocks = []string{
	BlockTypeControl,
	BlockTypeInput,
	BlockTypeQuery,
	BlockTypeChart,
	BlockTypeCard,
	BlockTypeTable,
	BlockTypeHierarchy,
}

// ReferenceBlockTypes is a list of block types we store references for
var ReferenceBlocks = []string{
	BlockTypeMod,
	BlockTypeQuery,
	BlockTypeControl,
	BlockTypeBenchmark,
	BlockTypeDashboard,
	BlockTypeContainer,
	BlockTypeCard,
	BlockTypeChart,
	BlockTypeHierarchy,
	BlockTypeImage,
	BlockTypeInput,
	BlockTypeTable,
	BlockTypeText,
	BlockTypeParam}
