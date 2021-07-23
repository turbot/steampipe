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

// ModTreeItem must be implemented by elements of the control hierarchy
// i.e. Control and Benchmark
type ModTreeItem interface {
	AddParent(ModTreeItem) error
	AddChild(child ModTreeItem) error
	Name() string
	GetTitle() string
	GetDescription() string
	GetTags() map[string]string
	GetChildren() []ModTreeItem
	// Path returns an array of parents in the control hierarchy
	Path() []string
}

// HclResource must be implemented by resources defined in HCL
type HclResource interface {
	Name() string
	CtyValue() (cty.Value, error)
	OnDecoded(*hcl.Block) hcl.Diagnostics
	AddReference(reference string)
}

// ResourceWithMetadata must be implenented by resources which supports reflection metadata
type ResourceWithMetadata interface {
	Name() string
	GetMetadata() *ResourceMetadata
	SetMetadata(metadata *ResourceMetadata)
}
