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
