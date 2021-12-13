package mod_installer

import (
	"fmt"
	"strings"

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
	updated := installData.Updated.FlatNames()
	updateCount := len(updated)
	if updateCount == 0 {
		return "All mods are up to date"
	}

	verb := getVerb(VerbUpdated)
	return fmt.Sprintf("\n%s %d %s:\n\t%s\n", verb, updateCount, utils.Pluralize("mod", updateCount), strings.Join(updated, "\n\t"))
}

func BuildInstallSummary(installData *InstallData) string {
	installed := installData.RecentlyInstalled.FlatNames()
	installCount := len(installed)
	if installCount == 0 {
		return "All mods are up to date"
	}

	verb := getVerb(VerbInstalled)
	return fmt.Sprintf("\n%s %d %s:\n\t%s\n", verb, installCount, utils.Pluralize("mod", installCount), strings.Join(installed, "\n\t"))
}

func BuildUninstallSummary(installData *InstallData) string {

	verb := getVerb(VerbUninstalled)
	installed := installData.Uninstalled.FlatNames()
	installCount := len(installed)

	if installCount == 0 {
		return "Nothing to uninstall"
	}
	return fmt.Sprintf("\n%s %d %s:\n\t%s\n", verb, installCount, utils.Pluralize("mod", installCount), strings.Join(installed, "\n\t"))
}

func BuildPruneSummary(pruned version_map.VersionListMap) string {
	pruneCount := len(pruned.FlatMap())

	verb := getVerb(VerbPruned)
	return fmt.Sprintf("\n%s %d %s:\n", verb, pruneCount, utils.Pluralize("mod", pruneCount))
}

//
//func BuildGetSummary(installData *InstallData, requiredVersions version_map.VersionConstraintMap) string {
//	// for every required version, see whether we inmstalled it or if it was already installed
//	var installed, alreadyInstalled []string
//
//	for name, versionConstraint := range requiredVersions {
//		if resolvedVersions, ok := installData.RecentlyInstalled[name]; ok {
//			for _, v := range resolvedVersions {
//				if v.Constraint == versionConstraint.Constraint.Original {
//					installed = append(installed, modconfig.ModVersionFullName(name, v.Version))
//					break
//				}
//			}
//		} else if resolvedVersions, ok := installData.AlreadyInstalled[name]; ok {
//			for _, v := range resolvedVersions {
//				if v.Constraint == versionConstraint.Constraint.Original {
//					alreadyInstalled = append(alreadyInstalled, modconfig.ModVersionFullName(name, v.Version))
//					break
//				}
//			}
//		}
//	}
//	res := ""
//	if installCount := len(installed); installCount > 0 {
//		res = fmt.Sprintf("\nInstalled %d %s:\n\t%s\n", installCount, utils.Pluralize("mod", installCount), strings.Join(installed, "\n\t"))
//	}
//	if len(alreadyInstalled) > 0 {
//		res += fmt.Sprintf("\nAlready installed:\n\t%s\n", strings.Join(alreadyInstalled, "\n\t"))
//	}
//
//	return res
//}
//
//func BuildUpdateSummary(installData *InstallData) string {
//	updated := installData.RecentlyInstalled.FlatNames()
//	if len(updated) == 0 {
//		return "All mods are up to date\n"
//	}
//
//	return fmt.Sprintf("\nUpdated %d %s:\n\t%s\n", len(updated), utils.Pluralize("update", len(updated)), strings.Join(updated, "\n\t"))
//}
//
//func BuildAvailableUpdateSummary(current, updates version_map.DependencyVersionMap) string {
//	if len(updates) == 0 {
//		return "No updated mods available"
//	}
//
//	// TODO
//	updateCount := 0
//	var strs []string
//	//for parent, deps := range updates {
//	//	strs = append(strs, fmt.Sprintf("required by %s:", parent))
//	//	for name, update := range deps {
//	//		// get the current installed version
//	//		currentDep := current[parent][name]
//	//		strs = append(strs, fmt.Sprintf("\tmod: %s, version constraint %s, currently installed %s, available %s", name, update.Constraint, currentDep.Version, update.Version))
//	//		updateCount++
//	//	}
//	//}
//
//	return strings.Join(append([]string{fmt.Sprintf("%d %s found:", updateCount, utils.Pluralize("update", updateCount))}, strs...), "\n")
//}
