package task

import (
	"fmt"
	"sort"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/plugin"
)

// check if there is a new version
func checkPluginVersions(installationID string) []string {
	var notificationLines []string
	if !viper.GetBool(constants.ArgUpdateCheck) {
		return notificationLines
	}

	updateReport := plugin.GetAllUpdateReport(installationID)

	var pluginsToUpdate []plugin.VersionCheckReport

	for _, r := range updateReport {
		if r.CheckResponse.Digest != r.Plugin.ImageDigest {
			pluginsToUpdate = append(pluginsToUpdate, r)
		}
	}

	if len(pluginsToUpdate) > 0 {
		notificationLines = pluginNotificationMessage(pluginsToUpdate)
	}
	return notificationLines
}

func pluginNotificationMessage(reports []plugin.VersionCheckReport) []string {
	var notificationLines = []string{
		"",
		"Updated versions of the following plugins are available:",
		"",
	}
	longestNameLength := 0
	for _, report := range reports {
		thisName := report.ShortName()
		if len(thisName) > longestNameLength {
			longestNameLength = len(thisName)
		}
	}

	// sort alphabetically
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].ShortName() < reports[j].ShortName()
	})

	for _, report := range reports {
		thisName := report.ShortName()
		line := ""
		if len(report.Plugin.Version) == 0 {
			format := fmt.Sprintf("  %%-%ds @ %%-10s  →  %%10s", longestNameLength)
			line = fmt.Sprintf(
				format,
				thisName,
				report.CheckResponse.Stream,
				constants.Bold(report.CheckResponse.Version),
			)
		} else {
			format := fmt.Sprintf("  %%-%ds @ %%-10s       %%10s → %%-10s", longestNameLength)
			line = fmt.Sprintf(
				format,
				thisName,
				report.CheckResponse.Stream,
				constants.Bold(report.Plugin.Version),
				constants.Bold(report.CheckResponse.Version),
			)
		}
		notificationLines = append(notificationLines, line)
	}
	notificationLines = append(notificationLines, "")
	notificationLines = append(notificationLines, fmt.Sprintf("You can update by running %s", constants.Bold("steampipe plugin update --all")))
	notificationLines = append(notificationLines, "")

	return notificationLines
}

// func getNameFromReport(report plugin.VersionCheckReport) string {
// 	return fmt.Sprintf("%s/%s", report.CheckResponse.Org, report.CheckResponse.Name)
// }
