package mod_installer

import "github.com/turbot/steampipe/steampipeconfig/modconfig"

type InstallOpts struct {
	WorkspacePath string
	ShouldUpdate  bool
	GetMods       map[string]*modconfig.ModVersionConstraint
	UpdateMods    map[string]*modconfig.ModVersionConstraint
}
