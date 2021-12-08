package mod_installer

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/turbot/steampipe/utils"
)

func BuildInstallSummary(installData *InstallData) string {
	var installedString, alreadyInstalledString string

	installed := installData.RecentlyInstalled.Flat()
	if installCount := len(installed); installCount > 0 {
		installedString = fmt.Sprintf("\nInstalled %d %s:\n\t%s\n", installCount, utils.Pluralize("mod", installCount), strings.Join(installed, "\n\t"))
	}
	res := fmt.Sprintf("%s%s\n", installedString, alreadyInstalledString)
	return res
}

func BuildGetSummary(installData *InstallData, requiredVersions modconfig.VersionConstraintMap) string {
	// for every required version, see whether we inmstalled it or if it was already installed
	var installed, alreadyInstalled []string

	for name, versionConstraint := range requiredVersions {
		if resolvedVersions, ok := installData.RecentlyInstalled[name]; ok {
			for _, v := range resolvedVersions {
				if v.Constraint == versionConstraint.Constraint.Original {
					installed = append(installed, modVersionFullName(name, v.Version))
					break
				}
			}
		} else if resolvedVersions, ok := installData.AlreadyInstalled[name]; ok {
			for _, v := range resolvedVersions {
				if v.Constraint == versionConstraint.Constraint.Original {
					alreadyInstalled = append(alreadyInstalled, modVersionFullName(name, v.Version))
					break
				}
			}
		}
	}
	res := ""
	if installCount := len(installed); installCount > 0 {
		res = fmt.Sprintf("\nInstalled %d %s:\n\t%s\n", installCount, utils.Pluralize("mod", installCount), strings.Join(installed, "\n\t"))
	}
	if len(alreadyInstalled) > 0 {
		res += fmt.Sprintf("\nAlready installed:\n\t%s\n", strings.Join(alreadyInstalled, "\n\t"))
	}

	return res
}

func BuildUpdateSummary(installData *InstallData) string {
	updated := installData.RecentlyInstalled.Flat()
	if len(updated) == 0 {
		return "All mods are up to date\n"
	}

	return fmt.Sprintf("\nUpdated %d %s:\n\t%s\n", len(updated), utils.Pluralize("update", len(updated)), strings.Join(updated, "\n\t"))
}

func BuildAvailableUpdateSummary(current, updates modconfig.WorkspaceLock) string {
	if len(updates) == 0 {
		return "No updated mods available"
	}

	updateCount := 0
	var strs []string
	for parent, deps := range updates {
		strs = append(strs, fmt.Sprintf("required by %s:", parent))
		for name, update := range deps {
			// get the current installed version
			currentDep := current[parent][name]
			strs = append(strs, fmt.Sprintf("\tmod: %s, version constraint %s, currently installed %s, available %s", name, update.Constraint, currentDep.Version, update.Version))
			updateCount++
		}
	}

	return strings.Join(append([]string{fmt.Sprintf("%d %s found:", updateCount, utils.Pluralize("update", updateCount))}, strs...), "\n")
}
