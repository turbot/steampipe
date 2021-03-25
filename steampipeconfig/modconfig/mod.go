package modconfig

import (
	"fmt"
	"strings"
)

type Mod struct {
	Name          string
	Title         string `hcl:"title"`
	Description   string `hcl:"description"`
	Version       string `hcl:"version"`
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

	return fmt.Sprintf(`Name: %s
Title: %s
Description: %s
Version: %s
Mod Dependencies: %s
Plugin Dependencies: %s
Queries: 
%s`,
		m.Name,
		m.Title,
		m.Description,
		m.Version,
		modDependStr,
		pluginDependStr,
		strings.Join(queryStrings, "\n"),
	)
}
