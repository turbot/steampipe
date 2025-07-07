package task

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
)

const (
	AvailableVersionsCacheStructVersion = 20230117
)

func (r *Runner) saveAvailableVersions(cli *CLIVersionCheckResponse, plugin map[string]plugin.PluginVersionCheckReport) error {
	utils.LogTime("Runner.saveAvailableVersions start")
	defer utils.LogTime("Runner.saveAvailableVersions end")

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

func (r *Runner) loadCachedVersions() (*AvailableVersionCache, error) {
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

	ctx := cmd.Context()

	if !showNotificationsForCommand(cmd, cmdArgs) {
		// do not do anything - just return
		return nil
	}

	if !r.hasAvailableVersion() {
		// nothing to display
		return nil
	}

	cachedVersions, err := r.loadCachedVersions()
	if err != nil {
		return err
	}

	tableBuffer, err := cachedVersions.asTable(ctx)
	if err != nil {
		return err
	}

	// table can be nil if there are no notifications to display
	if tableBuffer != nil {
		fmt.Println()            //nolint:forbidigo // acceptable
		fmt.Println(tableBuffer) //nolint:forbidigo // acceptable
	}

	return nil
}
