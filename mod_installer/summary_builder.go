package mod_installer

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/steampipeconfig/version_map"
	"github.com/turbot/steampipe/utils"
)

const (
	VerbInstalled   = "Installed"
	VerbUninstalled = "Uninstalled"
	VerbUpdated     = "Updated"
	VerbPruned      = "Pruned"
)

var dryRunVerbs = map[string]string{
	VerbInstalled:   "Would install",
	VerbUninstalled: "Would uninstall",
	VerbUpdated:     "Would update",
	VerbPruned:      "Would prune",
}

func getVerb(verb string) string {
	if viper.GetBool(constants.ArgDryRun) {
		verb = dryRunVerbs[verb]
	}
	return verb
}

func BuildUpdateSummary(installData *InstallData) string {
	// for now treat an install as update - we only install deps which are in the mod.sp but missing in the mod folder
	updateCount := len(installData.Updated.FlatMap())
	if updateCount == 0 {
		if len(installData.Lock.InstallCache) == 0 {
			return "No mods installed"
		}
		return "All mods are up to date"
	}
	updatedTree := installData.GetUpdatedTree()

	verb := getVerb(VerbUpdated)
	return fmt.Sprintf("\n%s %d %s:\n\n%s", verb, updateCount, utils.Pluralize("mod", updateCount), updatedTree.String())
}

func BuildInstallSummary(installData *InstallData) string {
	// for now treat an install as update - we only install deps which are in the mod.sp but missing in the mod folder
	installCount := len(installData.Installed.FlatMap())
	if installCount == 0 {
		if len(installData.Lock.InstallCache) == 0 {
			return "No mods installed"
		}
		return "All mods are up to date"
	}
	installedTree := installData.GetInstalledTree()

	verb := getVerb(VerbInstalled)
	return fmt.Sprintf("\n%s %d %s:\n\n%s", verb, installCount, utils.Pluralize("mod", installCount), installedTree.String())
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

func BuildPruneSummary(pruned version_map.VersionListMap) string {
	pruneCount := len(pruned.FlatMap())

	verb := getVerb(VerbPruned)
	return fmt.Sprintf("\n%s %d %s:\n", verb, pruneCount, utils.Pluralize("mod", pruneCount))
}
