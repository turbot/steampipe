package workspace

import (
	"fmt"

	"github.com/turbot/steampipe/db/db_local"

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
	client                  db_common.Client
}

func NewConnectionWatcher(invoker constants.Invoker, errorHandler func(error)) (*ConnectionWatcher, error) {
	client, err := db_local.NewLocalClient(invoker)
	if err != nil {
		return nil, err
	}

	w := &ConnectionWatcher{
		client: client,
	}

	watcherOptions := &utils.WatcherOptions{
		Directories: []string{constants.ConfigDir()},
		Include:     filehelpers.InclusionsFromExtensions([]string{constants.ConfigExtension}),
		ListFlag:    filehelpers.FilesRecursive,

		OnChange: func(events []fsnotify.Event) {
			w.handleFileWatcherEvent(events)
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

func (w *ConnectionWatcher) handleFileWatcherEvent([]fsnotify.Event) {
	config, err := steampipeconfig.LoadConnectionConfig()
	if err != nil {
		fmt.Println()
		utils.ShowError(err)
		return
	}
	steampipeconfig.Config = config
	refreshResult := w.client.RefreshConnectionAndSearchPaths()
	if refreshResult.Error != nil {
		fmt.Println()
		utils.ShowError(refreshResult.Error)
		return
	}
	// display any refresh warnings
	refreshResult.ShowWarnings()
}

func (w *ConnectionWatcher) Close() {
	w.client.Close()
	w.watcher.Close()
}
