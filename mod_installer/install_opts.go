package mod_installer

import (
	"github.com/turbot/steampipe/steampipeconfig/version_map"
)

type InstallOpts struct {
	WorkspacePath string
	Updating      bool
	AddMods       version_map.VersionConstraintMap
	UpdateMods    version_map.VersionConstraintMap
}
