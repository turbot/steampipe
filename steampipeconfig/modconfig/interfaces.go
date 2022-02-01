package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// MappableResource must be implemented by resources which can be created
// directly from a content file (e.g. sql, markdown)
type MappableResource interface {
	// InitialiseFromFile creates a mappable resource from a file path
	// It returns the resource, and the raw file data
	InitialiseFromFile(modPath, filePath string) (MappableResource, []byte, error)
	Name() string
	SetMod(*Mod)
	GetMetadata() *ResourceMetadata
	SetMetadata(*ResourceMetadata)
}

// ModTreeItem must be implemented by elements of the mod resource hierarchy
// i.e. Control, Benchmark, Report
type ModTreeItem interface {
	AddParent(ModTreeItem) error
	GetChildren() []ModTreeItem
	Name() string
	GetTitle() string
	GetDescription() string
	GetTags() map[string]string
	// GetPaths returns an array resource paths
	GetPaths() []NodePath
	SetPaths()
	GetMod() *Mod
}

// HclResource must be implemented by resources defined in HCL
type HclResource interface {
	// implemented by HclResourceBase
	AddRuntimeDependencies(*RuntimeDependency)
	GetRuntimeDependencies() map[string]*RuntimeDependency

	Name() string
	CtyValue() (cty.Value, error)
	OnDecoded(*hcl.Block) hcl.Diagnostics
	AddReference(ref *ResourceReference)
	SetMod(*Mod)
	GetMod() *Mod
	GetDeclRange() *hcl.Range
}

// ResourceWithMetadata must be implemented by resources which supports reflection metadata
type ResourceWithMetadata interface {
	Name() string
	GetMetadata() *ResourceMetadata
	SetMetadata(metadata *ResourceMetadata)
}

// QueryProvider must be implemented by resources which supports prepared statements, i.e. Control and Query
type QueryProvider interface {
	Name() string
	GetModName() string
	GetParams() []*ParamDef
	GetSQL() string
	GetQuery() *Query
	GetPreparedStatementName() string
	SetArgs(args *QueryArgs)
	SetParams(params []*ParamDef)
}

// block types which implement QueryProvider
var QueryProviderBlocks = []string{
	BlockTypeControl,
	BlockTypeQuery,
	BlockTypeChart,
	BlockTypeCounter,
	BlockTypeTable,
	BlockTypeControl,
	BlockTypeControl,
	BlockTypeControl,
	BlockTypeControl,
}

// ReportingLeafNode must be implemented by resources may be a leaf node in the repoort execution tree
type ReportingLeafNode interface {
	Name() string
	GetUnqualifiedName() string
	GetTitle() string
	GetWidth() int
	GetPaths() []NodePath
	GetSQL() string
	GetRuntimeDependencies() map[string]*RuntimeDependency
}

type ResourceMapsProvider interface {
	GetResourceMaps() *WorkspaceResourceMaps
}
