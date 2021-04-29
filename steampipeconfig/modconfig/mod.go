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

type Mod struct {
	Name string `cty:"name"`

	// note these must be consistent with the attributes defined in 'modSchema'
	Color         *string   `cty:"color" column:"color" column_type:"text"`
	Description   *string   `cty:"description" column:"description" column_type:"text"`
	Documentation *string   `cty:"documentation" column:"documentation" column_type:"text"`
	Icon          *string   `cty:"icon" column:"icon" column_type:"text"`
	Labels        *[]string `cty:"labels" column:"labels" column_type:"text[]"`
	Title         *string   `cty:"title" column:"title" column_type:"text"`

	// TODO do we need this?
	Version *string
	//ModDepends    []*ModVersion
	//PluginDepends []*PluginDependency
	Queries       map[string]*Query
	Controls      map[string]*Control
	ControlGroups map[string]*ControlGroup
	OpenGraph     *OpenGraph
	ModPath       string
	DeclRange     hcl.Range

	children []ControlTreeItem
	metadata *ResourceMetadata
}

// Schema :: implementation of HclResource
func (m *Mod) Schema() *hcl.BodySchema {
	// todo this could be done automatically if we had a tag for block properties

	var attributes []hcl.AttributeSchema
	for attribute := range GetAttributeDetails(m) {
		attributes = append(attributes, hcl.AttributeSchema{Name: attribute})
	}
	return &hcl.BodySchema{
		Attributes: attributes,
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "requires"},
			{Type: "opengraph"},
		}}
}

func (m *Mod) CtyValue() (cty.Value, error) {
	return getCtyValue(m)
}

func NewMod(shortName, modPath string) *Mod {
	return &Mod{
		Name:          shortName,
		Queries:       make(map[string]*Query),
		Controls:      make(map[string]*Control),
		ControlGroups: make(map[string]*ControlGroup),
		ModPath:       modPath,
	}
}

func (m *Mod) String() string {
	if m == nil {
		return ""
	}
	//var modDependStr []string
	//for _, d := range m.ModDepends {
	//	modDependStr = append(modDependStr, d.String())
	//}
	//var pluginDependStr []string
	//for _, d := range m.PluginDepends {
	//	pluginDependStr = append(pluginDependStr, d.String())
	//}
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
		m.Name,
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
	if helpers.StringSliceContains(parent.Path(), item.FullName()) {
		return fmt.Errorf("cyclical dependency adding '%s' into control tree - parent '%s'", item.FullName(), parentName)
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
	if parsedName.Mod != "" && parsedName.Mod != m.Name {
		return nil, fmt.Errorf("cannot find item '%s' in mod '%s' - it is a child of mod '%s'", fullName, m.Name, parsedName.Mod)
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
		return nil, fmt.Errorf("cannot find item '%s' in mod '%s'", fullName, m.Name)
	}
	return item, nil
}

func (m *Mod) AddResource(item HclResource) bool {
	switch r := item.(type) {
	case *Query:
		name := r.FullName()
		// check for dupes
		if _, ok := m.Queries[name]; ok {
			return false
		}
		m.Queries[name] = r
		return true
	case *Control:
		name := r.FullName()
		// check for dupes
		if _, ok := m.Controls[name]; ok {
			return false
		}
		m.Controls[name] = r
		return true
	case *ControlGroup:
		name := r.FullName()
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

// FullName :: implementation of ControlTreeItem, HclResource
func (m *Mod) FullName() string {

	if m.Version == nil {
		return fmt.Sprintf("mod.%s", m.Name)
	}
	return fmt.Sprintf("mod.%s@%s", m.Name, types.SafeString(m.Version))
}

// Path :: implementation of ControlTreeItem
func (m *Mod) Path() []string {
	return []string{m.FullName()}
}

// AddPseudoResource :: add resource to parse results, if there is no resource of same name
func (m *Mod) AddPseudoResource(resource MappableResource) {
	switch r := resource.(type) {
	case *Query:
		// check there is not already a query with the same name
		if _, ok := m.Queries[r.Name]; !ok {
			m.Queries[r.Name] = r
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
