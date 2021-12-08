package mod_installer

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/turbot/steampipe/utils"
)

func BuildInstallSummary(data *InstallData) string {
	var installedString, alreadyInstalledString string

	if installCount := len(data.RecentlyInstalled); installCount > 0 {

		installedString = fmt.Sprintf("\nInstalled %d %s:\n\t%s\n", installCount, utils.Pluralize("mod", installCount), strings.Join(data.RecentlyInstalled, "\n\t"))
	}
	res := fmt.Sprintf("%s%s\n", installedString, alreadyInstalledString)
	return res
}
func BuildGetSummary(data *InstallData, versions map[string]*modconfig.ModVersionConstraint) string {
	var installedString, alreadyInstalledString string
	if installCount := len(data.RecentlyInstalled); installCount > 0 {
		installedString = fmt.Sprintf("\nInstalled %d %s:\n\t%s\n", installCount, utils.Pluralize("mod", installCount), strings.Join(data.RecentlyInstalled, "\n\t"))
	}
	res := fmt.Sprintf("%s%s\n", installedString, alreadyInstalledString)
	return res
}

func BuildUpdateSummary(data *InstallData) string {
	var updatedString, updateAlreadyInstalledString string
	if installCount := len(data.RecentlyInstalled); installCount > 0 {
		updatedString = fmt.Sprintf("\nInstalled %d %s:\n\t%s\n", installCount, utils.Pluralize("update", installCount), strings.Join(data.RecentlyInstalled, "\n\t"))
	}
	if len(data.AlreadyInstalled) > 0 {
		updateAlreadyInstalledString = fmt.Sprintf("\nAlready installed:\n\t%s\n", strings.Join(data.AlreadyInstalled, "\n\t"))
	}
	return fmt.Sprintf("%s%s\n", updatedString, updateAlreadyInstalledString)
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
