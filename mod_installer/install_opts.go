package mod_installer

import (
	"github.com/turbot/steampipe/steampipeconfig/version_map"
)

type InstallOpts struct {
	WorkspacePath string
	Updating      bool
	DryRun        bool
	ModArgs       version_map.VersionConstraintMap
}
