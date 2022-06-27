package modinstaller

import (
	"fmt"

	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/steampipeconfig/versionmap"
	"github.com/turbot/steampipe/pkg/utils"
)

func (i *ModInstaller) GetRequiredModVersionsFromArgs(modsArgs []string) (versionmap.VersionConstraintMap, error) {
	var errors []error
	mods := make(versionmap.VersionConstraintMap, len(modsArgs))
	for _, modArg := range modsArgs {
		// create mod version from arg
		modVersion, err := modconfig.NewModVersionConstraint(modArg)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		// if we are updating there are a few checks we need to make
		if i.updating() {
			modVersion, err = i.getUpdateVersion(modArg, modVersion)
			if err != nil {
				errors = append(errors, err)
				continue
			}
		}
		if i.uninstalling() {
			// it is not valid to specify a mod version for uninstall
			if modVersion.HasVersion() {
				errors = append(errors, fmt.Errorf("invalid arg '%s' - cannot specify a version when uninstalling", modArg))
				continue
			}
		}

		mods[modVersion.Name] = modVersion
	}
	if len(errors) > 0 {
		return nil, utils.CombineErrors(errors...)
	}
	return mods, nil
}

func (i *ModInstaller) getUpdateVersion(modArg string, modVersion *modconfig.ModVersionConstraint) (*modconfig.ModVersionConstraint, error) {
	// verify the mod is already installed
	if i.installData.Lock.GetMod(modVersion.Name, i.workspaceMod) == nil {
		return nil, fmt.Errorf("cannot update '%s' as it is not installed", modArg)
	}

	// find the current dependency with this mod name
	// - this is what we will be using, to ensure we keep the same version constraint
	currentDependency := i.workspaceMod.GetModDependency(modVersion.Name)
	if currentDependency == nil {
		return nil, fmt.Errorf("cannot update '%s' as it is not a dependency of this workspace", modArg)
	}

	// it is not valid to specify a mod version - we will set the constraint from the modfile
	if modVersion.HasVersion() {
		return nil, fmt.Errorf("invalid arg '%s' - cannot specify a version when updating", modArg)
	}
	return currentDependency, nil
}
