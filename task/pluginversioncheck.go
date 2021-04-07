package task

import (
	"fmt"
	"os"
	"sort"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/plugin"
)

// check if there is a new version
func checkPluginVersions(installationID string) {
	if !viper.GetBool(constants.ArgUpdateCheck) {
		return
	}

	updateReport := plugin.GetAllUpdateReport(installationID)

	var pluginsToUpdate []plugin.VersionCheckReport

	for _, r := range updateReport {
		if r.CheckResponse.Digest != r.Plugin.ImageDigest {
			pluginsToUpdate = append(pluginsToUpdate, r)
		}
	}

	if len(pluginsToUpdate) > 0 {
		showPluginUpdateNotification(pluginsToUpdate)
	}
}

func showPluginUpdateNotification(reports []plugin.VersionCheckReport) {
	var notificationLines = [][]string{
		{""},
		{"Updated versions of the following plugins are available:"},
		{""},
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
			format := fmt.Sprintf("  %%-%ds @ %%-10s       %%21s", longestNameLength)
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
		notificationLines = append(notificationLines, []string{line})
	}
	notificationLines = append(notificationLines, []string{""})
	notificationLines = append(notificationLines, []string{
		fmt.Sprintf("You can update by running %s", constants.Bold("steampipe plugin update --all")),
	})
	notificationLines = append(notificationLines, []string{""})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{})                // no headers
	table.SetAlignment(tablewriter.ALIGN_LEFT) // align to the left
	table.SetAutoWrapText(false)               // do not wrap the text
	table.SetBorder(true)                      // there needs to be a border to give the dialog feel
	table.AppendBulk(notificationLines)        // Add Bulk Data

	fmt.Println()
	table.Render()
	fmt.Println()
}

// func getNameFromReport(report plugin.VersionCheckReport) string {
// 	return fmt.Sprintf("%s/%s", report.CheckResponse.Org, report.CheckResponse.Name)
// }
