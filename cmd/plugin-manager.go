package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/connection_watcher"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/plugin_manager"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/utils"
)

func pluginManagerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "plugin-manager",
		Run:    runPluginManagerCmd,
		Hidden: true,
	}
	cmdconfig.OnCmd(cmd)

	return cmd
}

func runPluginManagerCmd(cmd *cobra.Command, args []string) {
	logger := createPluginManagerLog()

	log.Printf("[INFO] Starting plugin manager")
	// build config map
	steampipeConfig, err := steampipeconfig.LoadConnectionConfig()
	if err != nil {
		log.Printf("[WARN] failed to load connection config: %s", err.Error())
		os.Exit(1)
	}
	configMap := connection_watcher.NewConnectionConfigMap(steampipeConfig.Connections)
	log.Printf("[TRACE] loaded config map")

	pluginManager := plugin_manager.NewPluginManager(configMap, logger)
	connectionWatcher, err := connection_watcher.NewConnectionWatcher(pluginManager.SetConnectionConfigMap)
	if err != nil {
		log.Printf("[WARN] failed to create connection watcher: %s", err.Error())
		utils.ShowError(err)
		os.Exit(1)
	}

	// close the connection watcher
	defer connectionWatcher.Close()

	log.Printf("[TRACE] about to serve")
	pluginManager.Serve()
}

func createPluginManagerLog() hclog.Logger {
	logName := fmt.Sprintf("plugin-%s.log", time.Now().Format("2006-01-02"))
	logPath := filepath.Join(constants.LogDir(), logName)
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("failed to open plugin manager log file: %s\n", err.Error())
		os.Exit(3)
	}
	logger := logging.NewLogger(&hclog.LoggerOptions{Output: f})
	log.SetOutput(logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)
	return logger
}
