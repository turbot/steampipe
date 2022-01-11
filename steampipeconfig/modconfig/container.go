package modconfig

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// Container is a struct representing the Report resource
type Container struct {
	ShortName       string
	FullName        string `cty:"name"`
	UnqualifiedName string

	ChildNames       []NamedItem `cty:"children" hcl:"children,optional"`
	ChildNameStrings []string    `column:"children,jsonb"`

	Width *int `cty:"width" hcl:"width" column:"width,text"`

	Mod       *Mod `cty:"mod"`
	DeclRange hcl.Range
	Paths     []NodePath `column:"path,jsonb"`

	parents  []ModTreeItem
	children []ModTreeItem
	metadata *ResourceMetadata
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

// AddReference implements HclResource
func (p *Container) AddReference(*ResourceReference) {}

// SetMod implements HclResource
func (p *Container) SetMod(mod *Mod) {
	p.Mod = mod
	p.UnqualifiedName = p.FullName
	p.FullName = fmt.Sprintf("%s.%s", mod.ShortName, p.FullName)
}

// GetDeclRange implements HclResource
func (p *Container) GetDeclRange() *hcl.Range {
	return &p.DeclRange
}

// OnDecoded implements HclResource
func (p *Container) OnDecoded(block *hcl.Block) hcl.Diagnostics {
	var res hcl.Diagnostics
	if len(p.ChildNames) == 0 {
		return nil
	}

	// validate each child name appears only once
	nameMap := make(map[string]bool)
	p.ChildNameStrings = make([]string, len(p.ChildNames))
	for i, n := range p.ChildNames {
		if nameMap[n.Name] {
			res = append(res, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("container '%s' has duplicate child name '%s'", p.FullName, n.Name),
				Subject:  &block.DefRange})

			continue
		}
		p.ChildNameStrings[i] = n.Name
		nameMap[n.Name] = true
	}

	// in order to populate the children in the order specified, we create an empty array and populate by index in AddChild
	p.children = make([]ModTreeItem, len(p.ChildNameStrings))
	return res
}

// AddChild implements ModTreeItem
func (p *Container) AddChild(child ModTreeItem) error {
	// if ChildNames is NOT set, children must hav ebeen declared inline
	if len(p.ChildNames) == 0 {
		p.children = append(p.children, child)
		p.ChildNameStrings = append(p.ChildNameStrings, child.Name())
		return nil
	}

	// so a children property must have been populated

	// now find which position this child is in the array
	for i, name := range p.ChildNameStrings {
		if name == child.Name() {
			p.children[i] = child
			return nil
		}
	}

	return fmt.Errorf("container '%s' has no child '%s'", p.Name(), child.Name())
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
	return p.children
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
