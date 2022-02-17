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
	GetMetadata() *ResourceMetadata
	SetMetadata(*ResourceMetadata)
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
	OnDecoded(*hcl.Block) hcl.Diagnostics
	AddReference(ref *ResourceReference)
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
	SetArgs(args *QueryArgs)
	SetParams(params []*ParamDef)
	GetPreparedStatementName() string
	GetPreparedStatementExecuteSQL(args *QueryArgs) (string, error)
}

// ParameterisedDashboardNode must be implemented by resources has params and args
type ParameterisedDashboardNode interface {
	GetParams() []*ParamDef
	GetArgs() *QueryArgs
}

// DashboardLeafNode must be implemented by resources may be a leaf node in the dashboard execution tree
type DashboardLeafNode interface {
	Name() string
	GetUnqualifiedName() string
	GetTitle() string
	GetWidth() int
	GetPaths() []NodePath

	// implemented by DashboardLeafNodeBase

	AddRuntimeDependencies(*RuntimeDependency)
	GetRuntimeDependencies() map[string]*RuntimeDependency
}

type ResourceMapsProvider interface {
	GetResourceMaps() *WorkspaceResourceMaps
}

type UniqueNameProvider interface {
	GetUniqueName(string) string
}
