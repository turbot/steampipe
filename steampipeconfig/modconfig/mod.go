package modconfig

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/turbot/go-kit/helpers"

	"github.com/turbot/go-kit/types"
)

type Mod struct {
	ShortName *string

	// note these must be consistent with the attributes defined in 'modSchema'
	Color         *string   `column:"color" column_type:"text"`
	Description   *string   `column:"description" column_type:"text"`
	Documentation *string   `column:"documentation" column_type:"text"`
	Icon          *string   `column:"icon" column_type:"text"`
	Labels        *[]string `column:"labels" column_type:"text[]"`
	Title         *string   `column:"title" column_type:"text"`

	// TODO do we need this?
	Version *string

	ModDepends    []*ModVersion
	PluginDepends []*PluginDependency
	Queries       map[string]*Query
	Controls      map[string]*Control
	ControlGroups map[string]*ControlGroup
	OpenGraph     *OpenGraph

	// direct children in the control tree
	Children []ControlTreeItem
}

func (m *Mod) FullName() string {
	if m.Version == nil {
		return types.SafeString(m.Name)
	}
	return fmt.Sprintf("%s@%s", m.Name, types.SafeString(m.Version))
}

func (m *Mod) String() string {
	if m == nil {
		return ""
	}
	var modDependStr []string
	for _, d := range m.ModDepends {
		modDependStr = append(modDependStr, d.String())
	}
	var pluginDependStr []string
	for _, d := range m.PluginDepends {
		pluginDependStr = append(pluginDependStr, d.String())
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
Mod Dependencies: %s
Plugin Dependencies: %s
Queries: 
%s
Controls: 
%s
Control Groups: 
%s`,
		types.SafeString(m.Name),
		types.SafeString(m.Title),
		types.SafeString(m.Description),
		versionString,
		modDependStr,
		pluginDependStr,
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
	parentName := types.SafeString(item.GetParentName())
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
	name, err := ParseModResourceName(fullName)
	if err != nil {
		return nil, err
	}
	// this function only finds items in the current mod
	if name.Mod != "" && name.Mod != types.SafeString(m.ShortName) {
		return nil, fmt.Errorf("cannot find item '%s' in mod '%s' - it is a child of mod '%s'", fullName, types.SafeString(m.ShortName), name.Mod)
	}
	// does name include an item type
	if name.ItemType == "" {
		return nil, fmt.Errorf("name '%s' does not specify an itemr type", fullName)
	}

	// so this item either does not specify a mod or specifies this mod
	var item ControlTreeItem
	var found bool
	switch name.ItemType {
	case BlockTypeControl:
		item, found = m.Controls[name.Name]
	case BlockTypeControlGroup:
		item, found = m.ControlGroups[name.Name]
	default:
		return nil, fmt.Errorf("ControlTreeItemFromName called invalid item type; '%s'", name.ItemType)
	}
	if !found {
		return nil, fmt.Errorf("cannot find item '%s' in mod '%s'", fullName, types.SafeString(m.ShortName))
	}
	return item, nil
}

// AddChild  :: implementation of ControlTreeItem
func (m *Mod) AddChild(child ControlTreeItem) error {
	m.Children = append(m.Children, child)
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

// Name :: implementation of ControlTreeItem
// note - for mod, long name and short name are the same
func (m *Mod) Name() string {
	name := types.SafeString(m.ShortName)
	// TODO think about name formats
	if m.Version == nil {
		//return fmt.Sprintf("mod.%s", name)
		return name
	}
	return fmt.Sprintf("%s@%s", name, types.SafeString(m.Version))
	//return fmt.Sprintf("mod.%s@%s", name, types.SafeString(m.Version))
}

// Path :: implementation of ControlTreeItem
func (m *Mod) Path() []string {
	return []string{m.Name()}
}

func (m *Mod) AddQueries(queries map[string]*Query) {
	// add mod into the reflection data of each query
	for _, q := range queries {
		q.Metadata.SetMod(m)
	}
	m.Queries = queries
}

func (m *Mod) AddControls(controls map[string]*Control) {
	// add mod into the reflection data of each query
	for _, c := range controls {
		c.Metadata.SetMod(m)
	}
	m.Controls = controls
}

func (m *Mod) AddControlGroups(controlGroups map[string]*ControlGroup) {
	// add mod into the reflection data of each query
	for _, c := range controlGroups {
		c.Metadata.SetMod(m)
	}
	m.ControlGroups = controlGroups
}
