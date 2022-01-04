package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// Container is a struct representing the Report resource
type Container struct {
	FullName  string `cty:"name"`
	ShortName string

	Width      *int `cty:"width" column:"width,text"`
	Containers []*Container
	Panels     []*Panel

	DeclRange hcl.Range
	Mod       *Mod `cty:"mod"`

	Children []string   `column:"children,jsonb"`
	Paths    []NodePath `column:"path,jsonb"`

	parents         []ModTreeItem
	metadata        *ResourceMetadata
	UnqualifiedName string
}

func NewContainer(block *hcl.Block) *Container {
	container := &Container{
		ShortName:       block.Labels[0],
		FullName:        fmt.Sprintf("container.%s", block.Labels[0]),
		UnqualifiedName: fmt.Sprintf("container.%s", block.Labels[0]),
		DeclRange:       block.DefRange,
	}
	return container
}

// GetMod implements HclResource
func (p *Container) GetMod() *Mod {
	return p.Mod
}

// CtyValue implements HclResource
func (p *Container) CtyValue() (cty.Value, error) {
	return getCtyValue(p)
}

// Name implements HclResource
// return name in format: 'container.<shortName>'
func (p *Container) Name() string {
	return p.FullName
}

// OnDecoded implements HclResource
func (p *Container) OnDecoded(*hcl.Block) hcl.Diagnostics {
	p.setChildNames()
	return nil
}

// AddChild implements ModTreeItem
func (p *Container) AddChild(child ModTreeItem) error {
	switch c := child.(type) {
	case *Panel:
		// avoid duplicates
		if !p.containsPanel(c.Name()) {
			p.Panels = append(p.Panels, c)
		}
	case *Report:
		return fmt.Errorf("panels cannot contain reports")
	}
	return nil
}

// AddParent implements ModTreeItem
func (p *Container) AddParent(parent ModTreeItem) error {
	p.parents = append(p.parents, parent)
	return nil
}

// GetParents implements ModTreeItem
func (p *Container) GetParents() []ModTreeItem {
	return p.parents
}

// GetChildren implements ModTreeItem
func (p *Container) GetChildren() []ModTreeItem {
	children := make([]ModTreeItem, len(p.Panels))
	for i, p := range p.Panels {
		children[i] = p
	}
	return children
}

// GetTitle implements ModTreeItem
func (p *Container) GetTitle() string {
	return ""
}

// GetDescription implements ModTreeItem
func (p *Container) GetDescription() string {
	return ""
}

// GetTags implements ModTreeItem
func (p *Container) GetTags() map[string]string {
	return nil
}

// GetPaths implements ModTreeItem
func (p *Container) GetPaths() []NodePath {
	// lazy load
	if len(p.Paths) == 0 {
		p.SetPaths()
	}

	return p.Paths
}

// SetPaths implements ModTreeItem
func (p *Container) SetPaths() {
	for _, parent := range p.parents {
		for _, parentPath := range parent.GetPaths() {
			p.Paths = append(p.Paths, append(parentPath, p.Name()))
		}
	}
}

func (p *Container) setChildNames() {
	numChildren := len(p.Containers)
	if numChildren == 0 {
		return
	}
	// set children names
	p.Children = make([]string, numChildren)

	for i, p := range p.Containers {
		p.Children[i] = p.Name()
	}
}

func (p *Container) containsContainer(name string) bool {
	// does this child already exist
	for _, existingContainer := range p.Containers {
		if existingContainer.Name() == name {
			return true
		}
	}
	return false
}

func (p *Container) containsPanel(name string) bool {
	// does this child already exist
	for _, existingPanel := range p.Panels {
		if existingPanel.Name() == name {
			return true
		}
	}
	return false
}
