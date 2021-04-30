package modconfig

import (
	"github.com/zclconf/go-cty/cty"
)

// MappableResource :: a mod resource which can be created directly from a content file (e.g. sql, markdown)
// InitialiseFromFile :: initialise the mod resource from the file of the given path
// return the created resource, and the file data
type MappableResource interface {
	InitialiseFromFile(modPath, filePath string) (MappableResource, []byte, error)
	Name() string

	GetMetadata() *ResourceMetadata
	SetMetadata(*ResourceMetadata)
}

type ControlTreeItem interface {
	// GetParentName :: get the name of the parent of this item
	GetParentName() string
	// SetParent :: set the parent of this item
	SetParent(ControlTreeItem) error
	// AddChild :: add a child to the item
	AddChild(child ControlTreeItem) error
	// Name :: name in the format <type>.<name>
	Name() string
	// Path ::array of parents in the control hiearchy
	Path() []string
}

// HclResource :: a resource which is defined in HCL
// - must be implemented by resource with attributes to parse
type HclResource interface {
	Name() string
	CtyValue() (cty.Value, error)
}

// ResourceWithMetadata :: a resource which supports reflection metadata
type ResourceWithMetadata interface {
	Name() string
	GetMetadata() *ResourceMetadata
	SetMetadata(metadata *ResourceMetadata)
}
