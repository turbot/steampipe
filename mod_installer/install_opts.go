package mod_installer

import "github.com/turbot/steampipe/steampipeconfig/modconfig"

type InstallOpts struct {
	WorkspacePath string
	ShouldUpdate  bool
	GetMods       modconfig.VersionConstraintMap
	UpdateMods    modconfig.VersionConstraintMap
}
