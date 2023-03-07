package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants/runtime"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/logging"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/connectionwatcher"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"github.com/turbot/steampipe/pluginmanager_service"
)

func pluginManagerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "plugin-manager",
		Run:    runPluginManagerCmd,
		Hidden: true,
	}
	cmdconfig.OnCmd(cmd).AddStringFlag(constants.ArgAppName, "", "The app name to use for database connections")

	return cmd
}

func runPluginManagerCmd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	logger := createPluginManagerLog()

	log.Printf("[INFO] starting plugin manager")
	// build config map
	steampipeConfig, err := steampipeconfig.LoadConnectionConfig()
	if err != nil {
		log.Printf("[WARN] failed to load connection config: %s", err.Error())
		os.Exit(1)
	}

	// the CLI will pass the Postgress AppName to use for db clients - this is to ensure the CLI does not hold up
	// shutting down the DB because of connections we have open (but will close)
	if viper.IsSet(constants.ArgAppName) {
		runtime.PgClientAppName = viper.GetString(constants.ArgAppName)
	}

	configMap := connectionwatcher.NewConnectionConfigMap(steampipeConfig.Connections)
	log.Printf("[TRACE] loaded config map: %s", strings.Join(steampipeConfig.ConnectionNames(), ","))

	pluginManager, err := pluginmanager_service.NewPluginManager(configMap, logger)
	if err != nil {
		log.Printf("[WARN] failed to create plugin manager: %s", err.Error())
		os.Exit(1)
	}

	if shouldRunConnectionWatcher() {
		log.Printf("[INFO] starting connection watcher")

		connectionWatcher, err := connectionwatcher.NewConnectionWatcher(pluginManager.OnConnectionConfigChanged, pluginManager.OnSchemaChanged)
		if err != nil {
			log.Printf("[WARN] failed to create connection watcher: %s", err.Error())
			error_helpers.ShowError(ctx, err)
			os.Exit(1)
		}

		// close the connection watcher
		defer connectionWatcher.Close()
	}

	log.Printf("[TRACE] about to serve")
	pluginManager.Serve()
}

func shouldRunConnectionWatcher() bool {
	// if EnvConnectionWatcher is set, overwrite the value in DefaultConnectionOptions
	if envStr, ok := os.LookupEnv(constants.EnvConnectionWatcher); ok {
		if parsedEnv, err := types.ToBool(envStr); err == nil {
			return parsedEnv
		}
	}
	return true
}

func createPluginManagerLog() hclog.Logger {
	logName := fmt.Sprintf("plugin-%s.log", time.Now().Format("2006-01-02"))
	logPath := filepath.Join(filepaths.EnsureLogDir(), logName)
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("failed to open plugin manager log file: %s\n", err.Error())
		os.Exit(3)
	}
	logger := logging.NewLogger(&hclog.LoggerOptions{
		Output:     f,
		TimeFn:     func() time.Time { return time.Now().UTC() },
		TimeFormat: "2006-01-02 15:04:05.000 UTC",
	})
	log.SetOutput(logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)
	return logger
}
