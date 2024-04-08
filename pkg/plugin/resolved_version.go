package plugin

import "fmt"

type ResolvedPluginVersion struct {
	PluginName string
	Version    string
	Constraint string
}

func NewResolvedPluginVersion(pluginName string, version string, constraint string) ResolvedPluginVersion {
	return ResolvedPluginVersion{
		PluginName: pluginName,
		Version:    version,
		Constraint: constraint,
	}
}

// GetVersionTag returns the <PluginName>:<Version> (turbot/chaos:0.4.1)
func (r ResolvedPluginVersion) GetVersionTag() string {
	return fmt.Sprintf("%s:%s", r.PluginName, r.Version)
}
