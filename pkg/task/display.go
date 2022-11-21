package task

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/utils"
)

const (
	NotificationsStructVersion = 20221121
)

type Notifications struct {
	StructVersion           uint32   `json:"struct_version"`
	CLINotificationsLines   []string `json:"cli_notifications"`
	PluginNotificationLines []string `json:"plugin_notifications"`
}

func (r *Runner) saveNotifications(cliNotificationsLines, pluginNotificationLines []string) error {
	utils.LogTime("Runner.saveNotifications start")
	defer utils.LogTime("Runner.saveNotifications end")

	if len(cliNotificationsLines)+len(pluginNotificationLines) == 0 {
		// nothing to save
		return nil
	}

	notifs := &Notifications{
		StructVersion:           NotificationsStructVersion,
		CLINotificationsLines:   cliNotificationsLines,
		PluginNotificationLines: pluginNotificationLines,
	}
	if files.FileExists(filepaths.NotificationsFilePath()) {
		os.Remove(filepaths.NotificationsFilePath())
	}
	f, err := os.Create(filepaths.NotificationsFilePath())
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(f)
	return encoder.Encode(notifs)
}

func (r *Runner) hasNotifications() bool {
	utils.LogTime("Runner.hasNotifications start")
	defer utils.LogTime("Runner.hasNotifications end")
	return files.FileExists(filepaths.NotificationsFilePath())
}

func (r *Runner) displayNotifications(cmd *cobra.Command, cmdArgs []string) error {
	utils.LogTime("Runner.displayNotifications start")
	defer utils.LogTime("Runner.displayNotifications end")

	if isSilentCmd(cmd, cmdArgs) {
		// do not do anything
		// just return
		return nil
	}
	if !r.hasNotifications() {
		// nothing to display
		return nil
	}
	f, err := os.Open(filepaths.NotificationsFilePath())
	if err != nil {
		return err
	}
	defer func() {
		// close the open file reader
		f.Close()
		// remove the file
		os.Remove(filepaths.NotificationsFilePath())
	}()

	notifications := &Notifications{}
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(notifications); err != nil {
		return err
	}

	notificationLines := append(notifications.CLINotificationsLines, notifications.PluginNotificationLines...)

	// convert notificationLines into an array of arrays
	// since that's what our table renderer expects
	var notificationTable = make([][]string, len(notificationLines))
	for i, line := range notificationLines {
		notificationTable[i] = []string{line}
	}

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
