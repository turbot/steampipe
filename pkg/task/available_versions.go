package task

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/pipe-fittings/v2/utils"
)

type AvailableVersionCache struct {
	StructVersion uint32                                     `json:"struct_version"`
	CliCache      *CLIVersionCheckResponse                   `json:"cli_version"`
	PluginCache   map[string]plugin.PluginVersionCheckReport `json:"plugin_version"`
}

func (av *AvailableVersionCache) asTable(ctx context.Context) (*bytes.Buffer, error) {
	notificationLines, err := av.buildNotification(ctx)
	if err != nil {
		return nil, err
	}
	notificationTable := utils.Map(notificationLines, func(line string) []string {
		return []string{line}
	})

	if len(notificationLines) == 0 {
		return nil, nil
	}

	// create a buffer writer to pass to the tablewriter
	// so that we can capture the output
	var buffer bytes.Buffer // c

	table := tablewriter.NewWriter(&buffer)
	table.SetHeader([]string{})                // no headers please
	table.SetAlignment(tablewriter.ALIGN_LEFT) // we align to the left
	table.SetAutoWrapText(false)               // let's not wrap the text
	table.SetBorder(true)                      // there needs to be a border to provide the dialog feel
	table.AppendBulk(notificationTable)        // Add Bulk Data

	// render the table into the buffer
	table.Render()
	return &buffer, nil
}

func (av *AvailableVersionCache) buildNotification(ctx context.Context) ([]string, error) {
	cliLines, err := av.cliNotificationMessage()
	if err != nil {
		return nil, err
	}
	pluginLines := av.pluginNotificationMessage(ctx)
	// convert notificationLines into an array of arrays
	// since that's what our table renderer expects
	return append(cliLines, pluginLines...), nil
}

func (av *AvailableVersionCache) cliNotificationMessage() ([]string, error) {
	info := av.CliCache
	if info == nil {
		return nil, nil
	}

	if info.NewVersion == "" {
		return nil, nil
	}

	newVersion, err := semver.NewVersion(info.NewVersion)
	if err != nil {
		return nil, err
	}

	currentVersion, err := semver.NewVersion(currentVersion)
	if err != nil {
		fmt.Println(fmt.Errorf("there's something wrong with the Current Version"))
		fmt.Println(err)
	}

	if newVersion.GreaterThan(currentVersion) {
		var downloadURLColor = color.New(color.FgYellow)
		var notificationLines = []string{
			"",
			fmt.Sprintf("A new version of Steampipe is available! %s → %s", constants.Bold(currentVersion), constants.Bold(newVersion)),
			fmt.Sprintf("You can update by downloading from %s", downloadURLColor.Sprint("https://steampipe.io/downloads")),
			"",
		}
		return notificationLines, nil
	}
	return nil, nil
}

func ppNotificationLines() []string {
	var notificationLines = []string{
		"",
		fmt.Sprintf("%s Steampipe mods and dashboards are now separately available in Powerpipe (%s), a new open-source project (%s).", constants.Bold("Introducing Powerpipe:"), color.YellowString("https://powerpipe.io"), color.YellowString("https://github.com/turbot/powerpipe")),
		"",
		fmt.Sprintf("The steampipe mod, check and dashboard commands %s in a future version. Migration guide - %s", constants.Bold("will be removed"), color.YellowString("https://powerpipe.io/blog/migrating-from-steampipe")),
		"",
	}
	return notificationLines
}

func (av *AvailableVersionCache) pluginNotificationMessage(ctx context.Context) []string {
	var pluginsToUpdate []plugin.PluginVersionCheckReport

	for _, r := range av.PluginCache {
		if plugin.UpdateRequired(r) {
			pluginsToUpdate = append(pluginsToUpdate, r)
		}
	}
	notificationLines := []string{}
	if len(pluginsToUpdate) > 0 {
		notificationLines = av.getPluginNotificationLines(pluginsToUpdate)
	}
	return notificationLines
}

func (av *AvailableVersionCache) getPluginNotificationLines(reports []plugin.PluginVersionCheckReport) []string {
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
				report.CheckResponse.Constraint,
				constants.Bold(report.CheckResponse.Version),
			)
		} else {
			version := report.CheckResponse.Version
			format := fmt.Sprintf("  %%-%ds @ %%-10s       %%10s → %%-10s", longestNameLength)
			// an arm64 binary of the plugin might exist for the same version
			if report.Plugin.Version == report.CheckResponse.Version {
				version = fmt.Sprintf("%s (arm64)", version)
			}
			line = fmt.Sprintf(
				format,
				thisName,
				report.CheckResponse.Constraint,
				constants.Bold(report.Plugin.Version),
				constants.Bold(version),
			)
		}
		notificationLines = append(notificationLines, line)
	}
	notificationLines = append(notificationLines, "")
	notificationLines = append(notificationLines, fmt.Sprintf("You can update by running %s", constants.Bold("steampipe plugin update --all")))
	notificationLines = append(notificationLines, "")

	return notificationLines
}

func ppNoptificationAsTable(colWidth int) (*tablewriter.Table, error) {
	notificationLines := ppNotificationLines()

	notificationTable := utils.Map(notificationLines, func(line string) []string {
		return []string{line}
	})

	if len(notificationLines) == 0 {
		return nil, nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{})                // no headers please
	table.SetAlignment(tablewriter.ALIGN_LEFT) // we align to the left
	table.SetAutoWrapText(true)                // let's wrap the text
	table.SetBorder(true)                      // there needs to be a border to provide the dialog feel
	table.SetColWidth(colWidth - 4)            // set the column width which matches the steampipe notification table width
	table.AppendBulk(notificationTable)        // Add Bulk Data

	return table, nil
}
