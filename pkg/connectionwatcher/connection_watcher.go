package connectionwatcher

import (
	"context"
	"log"

	"github.com/fsnotify/fsnotify"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/filewatcher"
	"github.com/turbot/go-kit/helpers"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
)

type ConnectionWatcher struct {
	fileWatcherErrorHandler   func(error)
	watcher                   *filewatcher.FileWatcher
	onConnectionConfigChanged func(configMap map[string]*sdkproto.ConnectionConfig)
}

func NewConnectionWatcher(onConnectionChanged func(configMap map[string]*sdkproto.ConnectionConfig)) (*ConnectionWatcher, error) {
	w := &ConnectionWatcher{
		onConnectionConfigChanged: onConnectionChanged,
	}

	watcherOptions := &filewatcher.WatcherOptions{
		Directories: []string{filepaths.EnsureConfigDir()},
		Include:     filehelpers.InclusionsFromExtensions([]string{constants.ConfigExtension}),
		ListFlag:    filehelpers.FilesRecursive,
		EventMask:   fsnotify.Create | fsnotify.Remove | fsnotify.Rename | fsnotify.Write,
		OnChange: func(events []fsnotify.Event) {
			w.handleFileWatcherEvent(events)
		},
	}
	watcher, err := filewatcher.NewWatcher(watcherOptions)
	if err != nil {
		return nil, err
	}
	w.watcher = watcher

	// set the file watcher error handler, which will get called when there are parsing errors
	// after a file watcher event
	w.fileWatcherErrorHandler = func(err error) {
		log.Printf("[WARN] failed to reload connection config: %s", err.Error())
	}

	watcher.Start()

	log.Printf("[INFO] created ConnectionWatcher")
	return w, nil
}

func (w *ConnectionWatcher) handleFileWatcherEvent(_ []fsnotify.Event) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[WARN] ConnectionWatcher caught a panic: %s", helpers.ToError(r).Error())
		}
	}()

	// this is a file system event handler and not bound to any context
	ctx := context.Background()

	log.Printf("[TRACE] ConnectionWatcher handleFileWatcherEvent")
	config, err := steampipeconfig.LoadConnectionConfig()
	if err != nil {
		log.Printf("[WARN] error loading updated connection config: %s", err.Error())
		return
	}
	log.Printf("[TRACE] loaded updated config")

	client, err := db_local.NewLocalClient(ctx, constants.InvokerConnectionWatcher, nil)
	if err != nil {
		log.Printf("[WARN] error creating client to handle updated connection config: %s", err.Error())
	}
	defer client.Close(ctx)

	log.Printf("[TRACE] loaded updated config")

	log.Printf("[TRACE] calling onConnectionConfigChanged")
	// convert config to format expected by plugin manager
	// (plugin manager cannot reference steampipe config to avoid circular deps)
	configMap := NewConnectionConfigMap(config.Connections)
	// call on changed callback
	// (this calls pluginmanager.SetConnectionConfigMap)
	w.onConnectionConfigChanged(configMap)

	log.Printf("[TRACE] calling RefreshConnectionAndSearchPaths")

	// We need to update the viper config and GlobalConfig
	// as these are both used by RefreshConnectionAndSearchPaths

	// set the global steampipe config
	steampipeconfig.GlobalConfig = config

	// now refresh connections and search paths
	refreshResult := client.RefreshConnectionAndSearchPaths(ctx)
	if refreshResult.Error != nil {
		log.Printf("[WARN] error refreshing connections: %s", refreshResult.Error)
		return
	}

	// display any refresh warnings
	refreshResult.ShowWarnings()
}

func (w *ConnectionWatcher) Close() {
	w.watcher.Close()
}
