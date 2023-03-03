package connectionwatcher

import (
	"context"
	"log"

	"github.com/fsnotify/fsnotify"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/filewatcher"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
)

type ConnectionChangedFunc func(configMap ConnectionConfigMap, refreshResult *steampipeconfig.RefreshConnectionResult)

type ConnectionWatcher struct {
	fileWatcherErrorHandler   func(error)
	watcher                   *filewatcher.FileWatcher
	onConnectionConfigChanged ConnectionChangedFunc
}

func NewConnectionWatcher(onConnectionChanged ConnectionChangedFunc) (*ConnectionWatcher, error) {
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

	log.Printf("[WARN] ConnectionWatcher handleFileWatcherEvent")
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

	log.Printf("[TRACE] calling RefreshConnectionAndSearchPaths")

	// We need to update the viper config and GlobalConfig
	// as these are both used by RefreshConnectionAndSearchPaths

	// set the global steampipe config
	steampipeconfig.GlobalConfig = config

	// The only configurations from GlobalConfig which have
	// impact during Refresh are Database options and the Connections
	// themselves.
	//
	// It is safe to ignore the Workspace Profile here since this
	// code runs in the plugin-manager and has been started with the
	// install-dir properly set from the active Workspace Profile
	//
	// Workspace Profile does not have any setting which can alter
	// behavior in service mode (namely search path). Therefore, it is safe
	// to use the GlobalConfig here and ignore Workspace Profile in general
	cmdconfig.SetDefaultsFromConfig(steampipeconfig.GlobalConfig.ConfigMap())

	// now refresh connections and search paths
	refreshResult := client.RefreshConnectionAndSearchPaths(ctx)
	if refreshResult.Error != nil {
		log.Printf("[WARN] error refreshing connections: %s", refreshResult.Error)
		return
	}
	// call on changed callback
	// convert config to format expected by plugin manager
	// (plugin manager cannot reference steampipe config to avoid circular deps)
	configMap := NewConnectionConfigMap(config.Connections)
	w.onConnectionConfigChanged(configMap, refreshResult)

	// display any refresh warnings
	// TODO send warnings on warning_stream (to FDW???)
	refreshResult.ShowWarnings()
	log.Printf("[WARN] File watch event done")
}

func (w *ConnectionWatcher) Close() {
	w.watcher.Close()
}
