package modconfig

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
