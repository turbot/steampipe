package task

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/turbot/steampipe/pluginmanager"
)

// check if there is a new version
func checkPluginVersions(installationID string) {
	if !shouldDoUpdateCheck() {
		return
	}

	updateReport := pluginmanager.GetPluginUpdateReport(installationID)

	pluginsToUpdate := []pluginmanager.VersionCheckReport{}

	for _, r := range updateReport {
		if r.CheckResponse.Digest != r.Plugin.ImageDigest {
			pluginsToUpdate = append(pluginsToUpdate, r)
		}
	}

	if len(pluginsToUpdate) > 0 {
		showPluginUpdateNotification(pluginsToUpdate)
	}
}

func showPluginUpdateNotification(reports []pluginmanager.VersionCheckReport) {
	var updateCmdColor = color.New(color.FgHiYellow, color.Bold)
	var oldVersionColor = color.New(color.FgHiRed, color.Bold)
	var newVersionColor = color.New(color.FgHiGreen, color.Bold)

	var notificationLines = [][]string{
		{""},
		{"Updated versions of the following plugins are available:"},
		{""},
	}
	for _, report := range reports {
		thisName := fmt.Sprintf("%s/%s", report.CheckResponse.Org, report.CheckResponse.Name)
		line := ""
		if len(report.Plugin.Version) == 0 {
			line = fmt.Sprintf(
				"%-20s @ %-10s %24s",
				thisName,
				report.CheckResponse.Stream,
				newVersionColor.Sprintf(report.CheckResponse.Version),
			)
		} else {
			line = fmt.Sprintf(
				"%-20s @ %-10s %10s → %-10s",
				thisName,
				report.CheckResponse.Stream,
				oldVersionColor.Sprintf(report.Plugin.Version),
				newVersionColor.Sprintf(report.CheckResponse.Version),
			)
		}
		notificationLines = append(notificationLines, []string{line})
	}
	notificationLines = append(notificationLines, []string{""})
	notificationLines = append(notificationLines, []string{
		fmt.Sprintf("You can update by running\n %60s", updateCmdColor.Sprintf("steampipe plugin update --all")),
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
