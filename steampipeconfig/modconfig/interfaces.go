package modconfig

import (
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

// ControlTreeItem must be implemented by elements of the control hieararchy
// i.e. Control and Benchmark
type ControlTreeItem interface {
	// SetParent sets the parent of this item
	SetParent(ControlTreeItem) error
	// AddChild adds a child to the item
	AddChild(child ControlTreeItem) error
	// Name returns the name in the format <type>.<name>
	Name() string
	// Path returns an array of parents in the control hiearchy
	Path() []string
}

// HclResource must be implemented by resources defined in HCL
type HclResource interface {
	Name() string
	CtyValue() (cty.Value, error)
	OnDecoded()
}

// ResourceWithMetadata must be implenented by resources which supports reflection metadata
type ResourceWithMetadata interface {
	Name() string
	GetMetadata() *ResourceMetadata
	SetMetadata(metadata *ResourceMetadata)
}
