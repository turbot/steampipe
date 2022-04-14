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
	GetUnqualifiedName() string
	GetMetadata() *ResourceMetadata
	SetMetadata(*ResourceMetadata)
	GetDeclRange() *hcl.Range
}

// ModTreeItem must be implemented by elements of the mod resource hierarchy
// i.e. Control, Benchmark, Dashboard
type ModTreeItem interface {
	AddParent(ModTreeItem) error
	GetChildren() []ModTreeItem
	Name() string
	GetUnqualifiedName() string
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
	Name() string
	GetUnqualifiedName() string
	CtyValue() (cty.Value, error)
	OnDecoded(*hcl.Block, ModResourcesProvider) hcl.Diagnostics
	AddReference(ref *ResourceReference)
	GetReferences() []*ResourceReference
	GetDeclRange() *hcl.Range
}

// ResourceWithMetadata must be implemented by resources which supports reflection metadata
type ResourceWithMetadata interface {
	Name() string
	GetMetadata() *ResourceMetadata
	SetMetadata(metadata *ResourceMetadata)
	SetAnonymous(block *hcl.Block)
	IsAnonymous() bool
}

// QueryProvider must be implemented by resources which supports prepared statements, i.e. Control and Query
type QueryProvider interface {
	Name() string
	GetArgs() *QueryArgs
	GetParams() []*ParamDef
	GetSQL() *string
	GetQuery() *Query
	SetArgs(*QueryArgs)
	SetParams([]*ParamDef)
	GetPreparedStatementName() string
	GetPreparedStatementExecuteSQL(*QueryArgs) (*ResolvedQuery, error)
	// implemented by QueryProviderBase
	AddRuntimeDependencies([]*RuntimeDependency)
	GetRuntimeDependencies() map[string]*RuntimeDependency
	RequiresExecution(QueryProvider) bool
	VerifyQuery(QueryProvider) error
}

// DashboardLeafNode must be implemented by resources may be a leaf node in the dashboard execution tree
type DashboardLeafNode interface {
	Name() string
	GetUnqualifiedName() string
	GetTitle() string
	GetDisplay() *string
	GetWidth() int
	GetPaths() []NodePath
	GetMetadata() *ResourceMetadata
}

type ModResourcesProvider interface {
	GetResourceMaps() *ModResources
}
