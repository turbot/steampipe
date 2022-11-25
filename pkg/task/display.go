package task

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/utils"
)

const (
	NotificationsStructVersion = 20221121
)

type Notifications struct {
	StructVersion      uint32   `json:"struct_version"`
	CliNotifications   []string `json:"cli_notifications"`
	PluginNotification []string `json:"plugin_notifications"`
}

func (n *Notifications) GetAll() []string {
	return append(n.CliNotifications, n.PluginNotification...)
}

func (r *Runner) saveNotifications(cliNotificationsLines, pluginNotificationLines []string) error {
	utils.LogTime("Runner.saveNotifications start")
	defer utils.LogTime("Runner.saveNotifications end")

	if len(cliNotificationsLines)+len(pluginNotificationLines) == 0 {
		// nothing to save
		return nil
	}

	notifs := &Notifications{
		StructVersion:      NotificationsStructVersion,
		CliNotifications:   cliNotificationsLines,
		PluginNotification: pluginNotificationLines,
	}
	// create the file - if it exists, it will be truncated by os.Create
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

func (r *Runner) getNotifications() (*Notifications, error) {
	utils.LogTime("Runner.getNotifications start")
	defer utils.LogTime("Runner.getNotifications end")
	f, err := os.Open(filepaths.NotificationsFilePath())
	if err != nil {
		return nil, err
	}
	notifications := &Notifications{}
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(notifications); err != nil {
		return nil, err
	}
	if err := error_helpers.CombineErrors(f.Close(), os.Remove(filepaths.NotificationsFilePath())); err != nil {
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
	if !r.hasNotifications() {
		// nothing to display
		return nil
	}

	notifications, err := r.getNotifications()
	if err != nil {
		return err
	}

	// convert notificationLines into an array of arrays
	// since that's what our table renderer expects
	var notificationTable = utils.Map(notifications.GetAll(), func(line string) []string {
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
