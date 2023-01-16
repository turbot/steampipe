package task

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"

	SemVer "github.com/Masterminds/semver"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/plugin"
	"github.com/turbot/steampipe/pkg/utils"
)

const (
	AvailableVersionsCacheStructVersion = 20221121
)

type AvailableVersionCache struct {
	StructVersion uint32                               `json:"struct_version"`
	CliCache      *CLIVersionCheckResponse             `json:"cli_version"`
	PluginCache   map[string]plugin.VersionCheckReport `json:"plugin_version"`
}

func (r *Runner) saveAvailableVersions(cli *CLIVersionCheckResponse, plugin map[string]plugin.VersionCheckReport) error {
	utils.LogTime("Runner.saveNotifications start")
	defer utils.LogTime("Runner.saveNotifications end")

	if cli == nil && len(plugin) == 0 {
		// nothing to save
		return nil
	}

	notifs := &AvailableVersionCache{
		StructVersion: AvailableVersionsCacheStructVersion,
		CliCache:      cli,
		PluginCache:   plugin,
	}
	// create the file - if it exists, it will be truncated by os.Create
	f, err := os.Create(filepaths.AvailableVersionsFilePath())
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	return encoder.Encode(notifs)
}

func (r *Runner) hasAvailableVersion() bool {
	utils.LogTime("Runner.hasNotifications start")
	defer utils.LogTime("Runner.hasNotifications end")
	return files.FileExists(filepaths.AvailableVersionsFilePath())
}

func (r *Runner) getAvailableVersions() (*AvailableVersionCache, error) {
	utils.LogTime("Runner.getNotifications start")
	defer utils.LogTime("Runner.getNotifications end")
	f, err := os.Open(filepaths.AvailableVersionsFilePath())
	if err != nil {
		return nil, err
	}
	notifications := &AvailableVersionCache{}
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(notifications); err != nil {
		return nil, err
	}
	if err := error_helpers.CombineErrors(f.Close(), os.Remove(filepaths.AvailableVersionsFilePath())); err != nil {
		// if Go couldn't close the file handle, no matter - this was just good practise
		// if Go couldn't remove the notification file, it'll get truncated next time we try to write to it
		// worst case is that the notification gets shown more than once
		log.Println("[TRACE] could not close/delete notification file", err)
	}
	return notifications, nil
}

// displayNotifications checks if there are any pending notifications to display
// and if so, displays them
// does nothing if the given command is a command where notifications are not displayed
func (r *Runner) displayNotifications(cmd *cobra.Command, cmdArgs []string) error {
	utils.LogTime("Runner.displayNotifications start")
	defer utils.LogTime("Runner.displayNotifications end")

	if !showNotificationsForCommand(cmd, cmdArgs) {
		// do not do anything - just return
		return nil
	}

	if !r.hasAvailableVersion() {
		// nothing to display
		return nil
	}

	availableVersions, err := r.getAvailableVersions()
	if err != nil {
		return err
	}

	cliLines, err := cliNotificationMessage(availableVersions.CliCache)
	if err != nil {
		return err
	}

	pluginLines := pluginNotificationMessage(availableVersions.PluginCache)

	// convert notificationLines into an array of arrays
	// since that's what our table renderer expects
	var notificationTable = utils.Map(append(cliLines, pluginLines...), func(line string) []string {
		return []string{line}
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{})                // no headers please
	table.SetAlignment(tablewriter.ALIGN_LEFT) // we align to the left
	table.SetAutoWrapText(false)               // let's not wrap the text
	table.SetBorder(true)                      // there needs to be a border to provide the dialog feel
	table.AppendBulk(notificationTable)        // Add Bulk Data

	fmt.Println()
	table.Render()
	fmt.Println()

	return nil
}

func cliNotificationMessage(info *CLIVersionCheckResponse) ([]string, error) {
	if info == nil {
		return nil, nil
	}

	if info.NewVersion == "" {
		return nil, nil
	}

	newVersion, err := SemVer.NewVersion(info.NewVersion)
	if err != nil {
		return nil, err
	}
	currentVersion, err := SemVer.NewVersion(currentVersion)

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

func pluginNotificationMessage(updateReport map[string]plugin.VersionCheckReport) []string {
	var pluginsToUpdate []plugin.VersionCheckReport

	for _, r := range updateReport {
		if skip, _ := plugin.SkipUpdate(r); !skip {
			pluginsToUpdate = append(pluginsToUpdate, r)
		}
	}
	notificationLines := []string{}
	if len(pluginsToUpdate) > 0 {
		notificationLines = getPluginNotificationLines(pluginsToUpdate)
	}
	return notificationLines
}

func getPluginNotificationLines(reports []plugin.VersionCheckReport) []string {
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
			version := report.CheckResponse.Version
			format := fmt.Sprintf("  %%-%ds @ %%-10s       %%10s → %%-10s", longestNameLength)
			// an arm64 binary of the plugin might exist for the same version
			if report.Plugin.Version == report.CheckResponse.Version {
				version = fmt.Sprintf("%s (arm64)", version)
			}
			line = fmt.Sprintf(
				format,
				thisName,
				report.CheckResponse.Stream,
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
