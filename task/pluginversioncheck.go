package task

import (
	"fmt"
	"os"

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
		thisName := fmt.Sprintf("%s/%s", report.CheckResponse.Org, report.CheckResponse.Name)
		if len(thisName) > longestNameLength {
			longestNameLength = len(thisName)
		}
	}
	for _, report := range reports {
		thisName := fmt.Sprintf("%s/%s", report.CheckResponse.Org, report.CheckResponse.Name)
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
	table.SetHeader([]string{})                // no headers please
	table.SetAlignment(tablewriter.ALIGN_LEFT) // we align to the left
	table.SetAutoWrapText(false)               // let's not wrap the text
	table.SetBorder(true)                      // there needs to be a border to give the dialog feel
	table.AppendBulk(notificationLines)        // Add Bulk Data

	fmt.Println()
	table.Render()
	fmt.Println()
}
