package modconfig

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// MappableResource must be implemented by resources which can be created
// directly from a content file (e.g. sql)
type MappableResource interface {
	HclResource
	ResourceWithMetadata
	// InitialiseFromFile creates a mappable resource from a file path
	// It returns the resource, and the raw file data
	InitialiseFromFile(modPath, filePath string) (MappableResource, []byte, error)
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
	GetBase() HclResource
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
	AddRuntimeDependencies([]*RuntimeDependency)
	GetRuntimeDependencies() map[string]*RuntimeDependency
}

type WithProvider interface {
	AddWith(with *DashboardWith) hcl.Diagnostics
	GetWiths() []*DashboardWith
	GetWith(string) (*DashboardWith, bool)
}

// QueryProvider must be implemented by resources which have query/sql
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
	ArgsInheritedFromBase() bool
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
	GetResource(parsedName *ParsedResourceName) (resource HclResource, found bool)
}

// NodeAndEdgeProvider must be implemented by any dashboard leaf node which supports edges and nodes
// (DashboardGraph, DashboardFlow, DashboardHierarchy)
// TODO [node_reuse] add NodeAndEdgeProviderImpl https://github.com/turbot/steampipe/issues/2918
type NodeAndEdgeProvider interface {
	QueryProvider
	WithProvider
	GetEdges() DashboardEdgeList
	SetEdges(DashboardEdgeList)
	GetNodes() DashboardNodeList
	SetNodes(DashboardNodeList)
	AddCategory(category *DashboardCategory) hcl.Diagnostics
	AddChild(child HclResource) hcl.Diagnostics
}
