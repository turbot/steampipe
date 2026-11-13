package connection

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/pipe-fittings/v2/filepaths"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

type recordingPluginManager struct {
	mockPluginManager
	configChanged bool
}

func (m *recordingPluginManager) OnConnectionConfigChanged(context.Context, ConnectionConfigMap, map[string]*plugin.Plugin) {
	m.configChanged = true
}

// The freeze in #5039 happens here: handleFileWatcherEvent returned before
// passing the new config on whenever any file in the folder failed to load.
func TestHandleFileWatcherEvent_BadFileDoesNotBlockOtherConnections(t *testing.T) {
	prevInstallDir := app_specific.InstallDir
	prevGlobalConfig := steampipeconfig.GlobalConfig
	app_specific.InstallDir = t.TempDir()
	t.Cleanup(func() {
		app_specific.InstallDir = prevInstallDir
		steampipeconfig.GlobalConfig = prevGlobalConfig
	})

	configDir := filepaths.EnsureConfigDir()
	files := map[string]string{
		"good.spc": "connection \"conn_good\" {\n  plugin = \"chaos\"\n}\n",
		"bad.spc":  "connection \"broken\" {\n  plugin = \"chaos\"\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(configDir, name), []byte(content), 0600); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	// the RefreshConnections goroutine this starts returns on the nil pool
	pm := &recordingPluginManager{}
	w := &ConnectionWatcher{pluginManager: pm}
	w.handleFileWatcherEvent(nil)

	if !pm.configChanged {
		t.Fatal("OnConnectionConfigChanged was not called: one bad file blocked every connection from syncing")
	}
	if _, ok := steampipeconfig.GlobalConfig.Connections["conn_good"]; !ok {
		t.Error("conn_good should be in the config the watcher applied")
	}
}
