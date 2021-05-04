package modconfig

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/go-kit/types"
)

// mod name used if a default mod is created for a workspace which does not define one explicitly
const defaultModName = "local"

type Mod struct {
	ShortName string `hcl:"name,label"`
	FullName  string `cty:"name"`

	// attributes
	Color         *string            `cty:"color" hcl:"color" column_type:"text"`
	Description   *string            `cty:"description" hcl:"description" column_type:"text"`
	Documentation *string            `cty:"documentation" hcl:"documentation" column_type:"text"`
	Icon          *string            `cty:"icon" hcl:"icon" column_type:"text"`
	Labels        *[]string          `cty:"labels" hcl:"labels"  column_type:"jsonb"`
	Tags          *map[string]string `cty:"tags" hcl:"tags" column_type:"jsonb"`
	Title         *string            `cty:"title" hcl:"title" column_type:"text"`

	// blocks
	Requires  *Requires  `hcl:"requires,block"`
	OpenGraph *OpenGraph `hcl:"opengraph,block"`

	// TODO do we need this?
	Version *string

	Queries       map[string]*Query
	Controls      map[string]*Control
	ControlGroups map[string]*ControlGroup
	ModPath       string
	DeclRange     hcl.Range

	children []ControlTreeItem
	metadata *ResourceMetadata
}

// Schema :: implementation of HclResource
func (m *Mod) Schema() *hcl.BodySchema {
	// todo this could be done fully generically if we had a tag for block properties
	schema := &hcl.BodySchema{Attributes: []hcl.AttributeSchema{
		{Name: "color"},
		{Name: "description"},
		{Name: "documentation"},
		{Name: "icon"},
		{Name: "labels"},
		{Name: "title"},
	}}
	schema.Blocks = []hcl.BlockHeaderSchema{
		{Type: BlockTypeRequires},
		{Type: BlockTypeOpengraph},
	}
	return schema

}

func (m *Mod) CtyValue() (cty.Value, error) {
	return getCtyValue(m)
}

func NewMod(shortName, modPath string, defRange hcl.Range) *Mod {
	return &Mod{
		ShortName:     shortName,
		FullName:      fmt.Sprintf("mod.%s", shortName),
		Queries:       make(map[string]*Query),
		Controls:      make(map[string]*Control),
		ControlGroups: make(map[string]*ControlGroup),
		ModPath:       modPath,
		DeclRange:     defRange,
	}
}

// CreateDefaultMod :: create a default mod created for a workspace with no mod definition
func CreateDefaultMod(modPath string) *Mod {
	return NewMod(defaultModName, modPath, hcl.Range{})
}

// IsDefaultMod :: is this mod a default mod created for a workspace with no mod definition
func (m *Mod) IsDefaultMod() bool {
	return m.ShortName == defaultModName
}

func (m *Mod) String() string {
	if m == nil {
		return ""
	}
	// build ordered list of query names
	var queryNames []string
	for name := range m.Queries {
		queryNames = append(queryNames, name)
	}
	sort.Strings(queryNames)

	var queryStrings []string
	for _, name := range queryNames {
		queryStrings = append(queryStrings, m.Queries[name].String())
	}
	// build ordered list of control names
	var controlNames []string
	for name := range m.Controls {
		controlNames = append(controlNames, name)
	}
	sort.Strings(controlNames)

	var controlStrings []string
	for _, name := range controlNames {
		controlStrings = append(controlStrings, m.Controls[name].String())
	}
	// build ordered list of control group names
	var controlGroupNames []string
	for name := range m.ControlGroups {
		controlGroupNames = append(controlGroupNames, name)
	}
	sort.Strings(controlGroupNames)

	var controlGroupStrings []string
	for _, name := range controlGroupNames {
		controlGroupStrings = append(controlGroupStrings, m.ControlGroups[name].String())
	}

	versionString := ""
	if m.Version != nil {
		versionString = fmt.Sprintf("\nVersion: %s", types.SafeString(m.Version))
	}
	return fmt.Sprintf(`Name: %s
Title: %s
Description: %s %s
//Mod Dependencies: %s
//Plugin Dependencies: %s
Queries: 
%s
Controls: 
%s
Control Groups: 
%s`,
		m.FullName,
		types.SafeString(m.Title),
		types.SafeString(m.Description),
		versionString,
		//modDependStr,
		//pluginDependStr,
		strings.Join(queryStrings, "\n"),
		strings.Join(controlStrings, "\n"),
		strings.Join(controlGroupStrings, "\n"),
	)
}

// IsControlTreeItem :: implementation of ControlTreeItem
// (mod is always top of the tree)
func (m *Mod) IsControlTreeItem() {}

func (m *Mod) BuildControlTree() error {
	for _, controlGroup := range m.ControlGroups {
		if err := m.addItemIntoControlTree(controlGroup); err != nil {
			return err
		}
	}
	for _, control := range m.Controls {
		if err := m.addItemIntoControlTree(control); err != nil {
			return err
		}
	}
	return nil
}

func (m *Mod) addItemIntoControlTree(item ControlTreeItem) error {
	parentName := item.GetParentName()
	var parent ControlTreeItem
	// if no parent is specified, the mod itself is the parent
	if parentName == "" {
		parent = m
	} else {
		// otherwise find parent
		var err error
		parent, err = m.ControlTreeItemFromName(parentName)
		if err != nil {
			return err
		}
	}

	// check this item does not exist in the parent path
	if helpers.StringSliceContains(parent.Path(), item.Name()) {
		return fmt.Errorf("cyclical dependency adding '%s' into control tree - parent '%s'", item.Name(), parentName)
	}
	// so we have a result - add into tree
	item.SetParent(parent)
	parent.AddChild(item)

	return nil
}

func (m *Mod) ControlTreeItemFromName(fullName string) (ControlTreeItem, error) {
	parsedName, err := ParseResourceName(fullName)
	if err != nil {
		return nil, err
	}
	// this function only finds items in the current mod
	if parsedName.Mod != "" && parsedName.Mod != m.ShortName {
		return nil, fmt.Errorf("cannot find item '%s' in mod '%s' - it is a child of mod '%s'", fullName, m.ShortName, parsedName.Mod)
	}
	// does name include an item type
	if parsedName.ItemType == "" {
		return nil, fmt.Errorf("name '%s' does not specify an item type", fullName)
	}

	// so this item either does not specify a mod or specifies this mod
	var item ControlTreeItem
	var found bool
	switch parsedName.ItemType {
	case BlockTypeControl:
		item, found = m.Controls[fullName]
	case BlockTypeControlGroup:
		item, found = m.ControlGroups[fullName]
	default:
		return nil, fmt.Errorf("ControlTreeItemFromName called invalid item type; '%s'", parsedName.ItemType)
	}
	if !found {
		return nil, fmt.Errorf("cannot find item '%s' in mod '%s'", fullName, m.ShortName)
	}
	return item, nil
}

func (m *Mod) AddResource(item HclResource) bool {
	switch r := item.(type) {
	case *Query:
		name := r.Name()
		// check for dupes
		if _, ok := m.Queries[name]; ok {
			return false
		}
		m.Queries[name] = r
		return true
	case *Control:
		name := r.Name()
		// check for dupes
		if _, ok := m.Controls[name]; ok {
			return false
		}
		m.Controls[name] = r
		return true
	case *ControlGroup:
		name := r.Name()
		// check for dupes
		if _, ok := m.ControlGroups[name]; ok {
			return false
		}
		m.ControlGroups[name] = r
		return true
	default:
		// mod does not store other resource types
		return true
	}
}

// AddChild  :: implementation of ControlTreeItem
func (m *Mod) AddChild(child ControlTreeItem) error {
	m.children = append(m.children, child)
	return nil
}

// GetParentName  :: implementation of ControlTreeItem
func (m *Mod) GetParentName() string {
	return ""
}

// SetParent :: implementation of ControlTreeItem
func (m *Mod) SetParent(ControlTreeItem) error {
	return errors.New("cannot set a parent on a mod")
}

// Name :: implementation of ControlTreeItem, HclResource
func (m *Mod) Name() string {

	if m.Version == nil {
		return m.FullName
	}
	return fmt.Sprintf("%s@%s", m.FullName, types.SafeString(m.Version))
}

// Path :: implementation of ControlTreeItem
func (m *Mod) Path() []string {
	return []string{m.Name()}
}

// AddPseudoResource :: add resource to parse results, if there is no resource of same name
func (m *Mod) AddPseudoResource(resource MappableResource) {
	switch r := resource.(type) {
	case *Query:
		// check there is not already a query with the same name
		if _, ok := m.Queries[r.ShortName]; !ok {
			m.Queries[r.ShortName] = r
			// set the mod on the query metadata
			r.GetMetadata().SetMod(m)
		}
	}
}

// GetMetadata :: implementation of HclResource
func (m *Mod) GetMetadata() *ResourceMetadata {
	return m.metadata
}

// SetMetadata :: implementation of HclResource
func (m *Mod) SetMetadata(metadata *ResourceMetadata) {
	m.metadata = metadata
}
