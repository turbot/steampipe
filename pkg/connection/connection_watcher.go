package connection

import (
	"context"
	"github.com/fsnotify/fsnotify"
	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/go-kit/filewatcher"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"log"
)

type ConnectionWatcher struct {
	fileWatcherErrorHandler   func(error)
	watcher                   *filewatcher.FileWatcher
	onConnectionConfigChanged func(ConnectionConfigMap)
}

func NewConnectionWatcher(onConnectionChanged func(ConnectionConfigMap)) (*ConnectionWatcher, error) {
	w := &ConnectionWatcher{
		onConnectionConfigChanged: onConnectionChanged,
	}

	watcherOptions := &filewatcher.WatcherOptions{
		Directories: []string{filepaths.EnsureConfigDir()},
		Include:     filehelpers.InclusionsFromExtensions([]string{constants.ConfigExtension}),
		ListFlag:    filehelpers.FilesRecursive,
		EventMask:   fsnotify.Create | fsnotify.Remove | fsnotify.Rename | fsnotify.Write | fsnotify.Chmod,
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

func (w *ConnectionWatcher) handleFileWatcherEvent(events []fsnotify.Event) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[WARN] ConnectionWatcher caught a panic: %s", helpers.ToError(r).Error())
		}
	}()
	// this is a file system event handler and not bound to any context
	ctx := context.Background()

	log.Printf("[INFO] ConnectionWatcher handleFileWatcherEvent")
	config, errorsAndWarnings := steampipeconfig.LoadConnectionConfig()
	if errorsAndWarnings.GetError() != nil {
		log.Printf("[WARN] error loading updated connection config: %v", errorsAndWarnings.GetError())
		return
	}

	log.Printf("[INFO] loaded updated config")

	// We need to update the viper config and GlobalConfig
	// as these are both used by RefreshConnectionAndSearchPathsWithLocalClient

	// set the global steampipe config
	steampipeconfig.GlobalConfig = config

	// call on changed callback - we must call this BEFORE calling refresh connections
	// convert config to format expected by plugin manager
	// (plugin manager cannot reference steampipe config to avoid circular deps)
	configMap := NewConnectionConfigMap(config.Connections)
	w.onConnectionConfigChanged(configMap)

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

	log.Printf("[INFO] calling RefreshConnections asyncronously")

	// call RefreshConnections asyncronously
	// the RefreshConnections implements its own locking to ensure only a singler execution and a single queues execution
	// TODO send warnings on warning_stream
	go RefreshConnections(ctx)

	log.Printf("[TRACE] File watch event done")
}

func (w *ConnectionWatcher) Close() {
	w.watcher.Close()
}
