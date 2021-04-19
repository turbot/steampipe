package modconfig

import (
	"fmt"
	"sort"
	"strings"

	"github.com/turbot/go-kit/types"
)

type Mod struct {
	Name          *string
	Title         *string `hcl:"title"`
	Description   *string `hcl:"description"`
	Version       *string
	ModDepends    []*ModVersion
	PluginDepends []*PluginDependency
	Queries       map[string]*Query
	Controls      map[string]*Control
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
%s`,
		types.SafeString(m.Name),
		types.SafeString(m.Title),
		types.SafeString(m.Description),
		versionString,
		modDependStr,
		pluginDependStr,
		strings.Join(queryStrings, "\n"),
		strings.Join(controlStrings, "\n"),
	)
}
