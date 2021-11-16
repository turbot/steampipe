package workspace

import (
	"fmt"

	"github.com/turbot/steampipe/plugin_manager"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"

	"github.com/turbot/steampipe/cmdconfig"
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
	// set the global steampipe config
	steampipeconfig.GlobalConfig = config
	// update the viper default based on this loaded config
	cmdconfig.SetViperDefaults(config)
	refreshResult := w.client.RefreshConnectionAndSearchPaths()
	if refreshResult.Error != nil {
		fmt.Println()
		utils.ShowError(refreshResult.Error)
		return
	}

	configMap := pb.NewConnectionConfigMap(steampipeconfig.GlobalConfig.Connections)
	pluginManager, err := plugin_manager.GetPluginManager()
	if err != nil {
		refreshResult.AddWarning(fmt.Sprintf("connection config watcher failed to connect to plugin manager: %s", err.Error()))
	} else {
		// so we got a plugin manager client - set connection config
		_, err := pluginManager.SetConnectionConfigMap(&pb.SetConnectionConfigMapRequest{ConfigMap: configMap})
		if err != nil {
			refreshResult.AddWarning(fmt.Sprintf("connection config watcher failed to set connection config in plugin manager: %s", err.Error()))
		}
	}
	// display any refresh warnings
	refreshResult.ShowWarnings()
}

func (w *ConnectionWatcher) Close() {
	w.client.Close()
	w.watcher.Close()
}
