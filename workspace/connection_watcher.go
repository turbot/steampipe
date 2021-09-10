package workspace

import (
	"fmt"

	"github.com/turbot/steampipe/steampipeconfig"

	"github.com/fsnotify/fsnotify"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/utils"
)

type ConnectionWatcher struct {
	fileWatcherErrorHandler func(error)
	watcherError            error
	watcher                 *utils.FileWatcher
}

func NewConnectionWatcher(client db_common.Client, errorHandler func(error)) (*ConnectionWatcher, error) {
	w := &ConnectionWatcher{}

	watcherOptions := &utils.WatcherOptions{
		Directories: []string{constants.ConfigDir()},
		Include:     filehelpers.InclusionsFromExtensions([]string{constants.ConfigExtension}),
		ListFlag:    filehelpers.FilesRecursive,

		OnChange: func(events []fsnotify.Event) {
			w.handleFileWatcherEvent(client, events)
		},
	}
	watcher, err := utils.NewWatcher(watcherOptions)
	if err != nil {
		return nil, err
	}
	w.watcher = watcher

	// set the file watcher error handler, which will get called when there are parsing errors
	// after a file watcher event
	w.fileWatcherErrorHandler = errorHandler
	if w.fileWatcherErrorHandler == nil {
		w.fileWatcherErrorHandler = func(err error) {
			fmt.Println()
			utils.ShowErrorWithMessage(err, "Failed to reload mod from file watcher")
		}
	}

	return w, nil
}

func (w *ConnectionWatcher) handleFileWatcherEvent(client db_common.Client, events []fsnotify.Event) {
	// TODO add new function to just load connection config, not options
	config, err := steampipeconfig.LoadSteampipeConfig("", "")
	if err != nil {
		fmt.Println()
		utils.ShowError(err)
		return
	}
	steampipeconfig.Config = config
	refreshResult := client.RefreshConnectionAndSearchPaths()
	if refreshResult.Error != nil {
		fmt.Println()
		utils.ShowError(refreshResult.Error)
		return
	}
	// display any refresh warnings
	refreshResult.ShowWarnings()
}

func (w *ConnectionWatcher) Close() {
	w.watcher.Close()
}
