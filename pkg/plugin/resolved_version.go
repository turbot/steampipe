package plugin

import "fmt"

type ResolvedPluginVersion struct {
	PluginName string
	Version    string // TODO: Should this be a semver.Version ?
	Constraint string
}

func NewResolvedPluginVersion(pluginName string, version string, constraint string) ResolvedPluginVersion {
	return ResolvedPluginVersion{
		PluginName: pluginName,
		Version:    version,
		Constraint: constraint,
	}
}

func (r ResolvedPluginVersion) GetVersionTag() string {
	return fmt.Sprintf("%s:%s", r.PluginName, r.Version)
}

func (r ResolvedPluginVersion) GetNameAndConstraint() string {
	return fmt.Sprintf("%s@%s", r.PluginName, r.Constraint)
}
