package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// MappableResource must be implemented by resources which can be created
// directly from a content file (e.g. sql)
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

// HclResource must be implemented by resources defined in HCL
type HclResource interface {
	// TODO  [node_reuse] rename to GetName/GetFullName
	Name() string
	GetTitle() string
	GetUnqualifiedName() string
	OnDecoded(*hcl.Block, ResourceMapsProvider) hcl.Diagnostics
	GetDeclRange() *hcl.Range
	BlockType() string
	GetDescription() string
	GetDocumentation() string
	GetTags() map[string]string
	SetTopLevel(bool)
	IsTopLevel() bool
	GetHclResourceImpl() *HclResourceImpl
}

// ModTreeItem must be implemented by elements of the mod resource hierarchy
// i.e. Control, Benchmark, Dashboard
type ModTreeItem interface {
	HclResource
	AddParent(ModTreeItem) error
	GetParents() []ModTreeItem
	GetChildren() []ModTreeItem
	// GetPaths returns an array resource paths
	GetPaths() []NodePath
	SetPaths()
	GetMod() *Mod
	GetModTreeItemImpl() *ModTreeItemImpl
}

// RuntimeDependencyProvider is implemented by all QueryProviders and Dashboard
type RuntimeDependencyProvider interface {
	ModTreeItem
	AddWith(with *DashboardWith) hcl.Diagnostics
	GetWiths() []*DashboardWith
	AddRuntimeDependencies([]*RuntimeDependency)
	GetRuntimeDependencies() map[string]*RuntimeDependency
}

// QueryProvider must be implemented by resources which supports prepared statements, i.e. Control and Query
type QueryProvider interface {
	RuntimeDependencyProvider
	GetArgs() *QueryArgs
	GetParams() []*ParamDef
	GetSQL() *string
	GetQuery() *Query
	SetArgs(*QueryArgs)
	SetParams([]*ParamDef)
	GetResolvedQuery(*QueryArgs) (*ResolvedQuery, error)
	RequiresExecution(QueryProvider) bool
	ValidateQuery() hcl.Diagnostics
	MergeParentArgs(QueryProvider, QueryProvider) hcl.Diagnostics
	GetQueryProviderImpl() *QueryProviderImpl
	ParamsInheritedFromBase() bool
}

type CtyValueProvider interface {
	CtyValue() (cty.Value, error)
}

// ResourceWithMetadata must be implemented by resources which supports reflection metadata
type ResourceWithMetadata interface {
	Name() string
	GetMetadata() *ResourceMetadata
	SetMetadata(metadata *ResourceMetadata)
	SetAnonymous(block *hcl.Block)
	IsAnonymous() bool
	AddReference(ref *ResourceReference)
	GetReferences() []*ResourceReference
}

// DashboardLeafNode must be implemented by resources may be a leaf node in the dashboard execution tree
type DashboardLeafNode interface {
	ModTreeItem
	ResourceWithMetadata
	GetDisplay() string
	GetType() string
	GetWidth() int
}

type ResourceMapsProvider interface {
	GetResourceMaps() *ResourceMaps
}

// NodeAndEdgeProvider must be implemented by any dashboard leaf node which supports edges and nodes
// (DashboardGraph, DashboardFlow, DashboardHierarchy)
type NodeAndEdgeProvider interface {
	QueryProvider
	GetEdges() DashboardEdgeList
	SetEdges(DashboardEdgeList)
	GetNodes() DashboardNodeList
	SetNodes(DashboardNodeList)
	AddCategory(category *DashboardCategory) hcl.Diagnostics
	AddChild(child HclResource) hcl.Diagnostics
}
