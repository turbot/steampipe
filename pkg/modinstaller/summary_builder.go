package modinstaller

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig/versionmap"
	"github.com/turbot/steampipe/pkg/utils"
)

const (
	VerbInstalled   = "Installed"
	VerbUninstalled = "Uninstalled"
	VerbUpgraded    = "Upgraded"
	VerbDowngraded  = "Downgraded"
	VerbPruned      = "Pruned"
)

var dryRunVerbs = map[string]string{
	VerbInstalled:   "Would install",
	VerbUninstalled: "Would uninstall",
	VerbUpgraded:    "Would upgrade",
	VerbDowngraded:  "Would downgrade",
	VerbPruned:      "Would prune",
}

func getVerb(verb string) string {
	if viper.GetBool(constants.ArgDryRun) {
		verb = dryRunVerbs[verb]
	}
	return verb
}

func BuildInstallSummary(installData *InstallData) string {
	// for now treat an install as update - we only install deps which are in the mod.sp but missing in the mod folder
	modDependencyPath := installData.WorkspaceMod.GetModDependencyPath()
	installCount, installedTreeString := getInstallationResultString(installData.Installed, modDependencyPath)
	uninstallCount, uninstalledTreeString := getInstallationResultString(installData.Uninstalled, modDependencyPath)
	upgradeCount, upgradeTreeString := getInstallationResultString(installData.Upgraded, modDependencyPath)
	downgradeCount, downgradeTreeString := getInstallationResultString(installData.Downgraded, modDependencyPath)

	var installString, upgradeString, downgradeString, uninstallString string
	if installCount > 0 {
		verb := getVerb(VerbInstalled)
		installString = fmt.Sprintf("\n%s %d %s:\n\n%s\n", verb, installCount, utils.Pluralize("mod", installCount), installedTreeString)
	}
	if uninstallCount > 0 {
		verb := getVerb(VerbUninstalled)
		uninstallString = fmt.Sprintf("\n%s %d %s:\n\n%s\n", verb, uninstallCount, utils.Pluralize("mod", uninstallCount), uninstalledTreeString)
	}
	if upgradeCount > 0 {
		verb := getVerb(VerbUpgraded)
		upgradeString = fmt.Sprintf("\n%s %d %s:\n\n%s\n", verb, upgradeCount, utils.Pluralize("mod", upgradeCount), upgradeTreeString)
	}
	if downgradeCount > 0 {
		verb := getVerb(VerbDowngraded)
		downgradeString = fmt.Sprintf("\n%s %d %s:\n\n%s\n", verb, downgradeCount, utils.Pluralize("mod", downgradeCount), downgradeTreeString)
	}

	if installCount+uninstallCount+upgradeCount+downgradeCount == 0 {
		if len(installData.Lock.InstallCache) == 0 {
			return "No mods are installed"
		}
		return "All mods are up to date"
	}
	return fmt.Sprintf("%s%s%s%s", installString, upgradeString, downgradeString, uninstallString)
}

func getInstallationResultString(items versionmap.DependencyVersionMap, modDependencyPath string) (int, string) {
	var res string
	count := len(items.FlatMap())
	if count > 0 {
		tree := items.GetDependencyTree(modDependencyPath)
		res = tree.String()
	}
	return count, res
}

func BuildUninstallSummary(installData *InstallData) string {
	// for now treat an install as update - we only install deps which are in the mod.sp but missing in the mod folder
	uninstallCount := len(installData.Uninstalled.FlatMap())
	if uninstallCount == 0 {
		return "Nothing uninstalled"
	}
	uninstalledTree := installData.GetUninstalledTree()

	verb := getVerb(VerbUninstalled)
	return fmt.Sprintf("\n%s %d %s:\n\n%s", verb, uninstallCount, utils.Pluralize("mod", uninstallCount), uninstalledTree.String())
}

func BuildPruneSummary(pruned versionmap.VersionListMap) string {
	pruneCount := len(pruned.FlatMap())

	verb := getVerb(VerbPruned)
	return fmt.Sprintf("\n%s %d %s:\n", verb, pruneCount, utils.Pluralize("mod", pruneCount))
}
