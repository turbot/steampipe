package mod_installer

import (
	"fmt"
	"strings"

	"github.com/turbot/steampipe/utils"
)

func BuildInstallSummary(data *InstallData) string {
	var installedString, alreadyInstalledString string
	if installCount := len(data.RecentlyInstalled); installCount > 0 {
		installedString = fmt.Sprintf("\nInstalled %d %s:\n\t%s\n", installCount, utils.Pluralize("mod", installCount), strings.Join(data.RecentlyInstalled, "\n\t"))
	}
	if len(data.AlreadyInstalled) > 0 {
		alreadyInstalledString = fmt.Sprintf("\nAlready installed:\n\t%s\n", strings.Join(data.AlreadyInstalled, "\n\t"))
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
