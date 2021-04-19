package modconfig

import (
	"fmt"
	"sort"
	"strings"
)

type Mod struct {
	Name          string
	Title         string `hcl:"title"`
	Description   string `hcl:"description"`
	Version       string
	ModDepends    []*ModVersion
	PluginDepends []*PluginDependency
	Queries       []*Query
}

func (m *Mod) FullName() string {
	if m.Version == "" {
		return m.Name
	}
	return fmt.Sprintf("%s@%s", m.Name, m.Version)
}

// PopulateQueries :: convert a map of queries into a sorted array
// and set Queries property
func (mod *Mod) PopulateQueries(queries map[string]*Query) {
	for _, q := range queries {
		mod.Queries = append(mod.Queries, q)
	}
	sort.Slice(mod.Queries, func(i, j int) bool {
		return mod.Queries[i].Name < mod.Queries[j].Name
	})
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
	var queryStrings []string
	for _, q := range m.Queries {
		queryStrings = append(queryStrings, q.String())
	}

	versionString := ""
	if m.Version != "" {
		versionString = fmt.Sprintf("\nVersion: %s", m.Version)
	}
	return fmt.Sprintf(`Name: %s
Title: %s
Description: %s %s
Mod Dependencies: %s
Plugin Dependencies: %s
Queries: 
%s`,
		m.Name,
		m.Title,
		m.Description,
		versionString,
		modDependStr,
		pluginDependStr,
		strings.Join(queryStrings, "\n"),
	)
}
