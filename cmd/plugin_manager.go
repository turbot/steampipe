package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/go-kit/logging"
	"github.com/turbot/go-kit/types"
	sdklogging "github.com/turbot/steampipe-plugin-sdk/v5/logging"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/v2/pkg/cmdconfig"
	"github.com/turbot/steampipe/v2/pkg/connection"
	"github.com/turbot/steampipe/v2/pkg/constants"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
	"github.com/turbot/steampipe/v2/pkg/pluginmanager_service"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
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

func runPluginManagerCmd(cmd *cobra.Command, _ []string) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = helpers.ToError(r)
		}
		if err != nil {
			// write to stdout so the plugin manager can extract the error message
			fmt.Println(fmt.Sprintf("%s%s", plugin.PluginStartupFailureMessage, err.Error()))
		}
		os.Exit(1)
	}()

	err = doRunPluginManager(cmd)
}

func doRunPluginManager(cmd *cobra.Command) error {
	pluginManager, err := createPluginManager(cmd)
	if err != nil {
		return err
	}

	if shouldRunConnectionWatcher() {
		log.Printf("[INFO] starting connection watcher")
		connectionWatcher, err := connection.NewConnectionWatcher(pluginManager)
		if err != nil {
			log.Printf("[ERROR] failed to create connection watcher: %v", err)
			return err
		}
		log.Printf("[INFO] connection watcher created successfully")

		// close the connection watcher
		defer connectionWatcher.Close()
	} else {
		log.Printf("[WARN] connection watcher is DISABLED")
	}

	log.Printf("[INFO] about to serve")
	pluginManager.Serve()
	return nil
}

func createPluginManager(cmd *cobra.Command) (*pluginmanager_service.PluginManager, error) {
	ctx := cmd.Context()
	logger := createPluginManagerLog()

	log.Printf("[INFO] starting plugin manager")
	// build config map
	steampipeConfig, errorsAndWarnings := steampipeconfig.LoadConnectionConfig(ctx)
	if errorsAndWarnings.GetError() != nil {
		log.Printf("[WARN] failed to load connection config: %v", errorsAndWarnings.GetError())
		return nil, errorsAndWarnings.Error
	}

	// add signal handler for sigpipe - this will be raised if we call displayWarning as stdout is piped
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGPIPE)
	go func() {
		for {
			// swallow signal
			<-signalCh
		}
	}()

	// create a map of connections configs, excluding connections in error
	configMap := connection.NewConnectionConfigMap(steampipeConfig.Connections)
	log.Printf("[TRACE] loaded config map: %s", strings.Join(steampipeConfig.ConnectionNames(), ","))

	pluginManager, err := pluginmanager_service.NewPluginManager(ctx, configMap, steampipeConfig.PluginsInstances, logger)
	if err != nil {
		log.Printf("[WARN] failed to create plugin manager: %s", err.Error())
		return nil, err
	}

	return pluginManager, nil
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
	// we use this logger to log from the plugin processes
	// the plugin processes uses the `EscapeNewlineWriter` to map the '\n' byte to "\n" string literal
	// this is to allow the plugin to send multiline log messages as a single log line.
	//
	// here we apply the reverse mapping to get back the original message
	writer := sdklogging.NewUnescapeNewlineWriter(logging.NewRotatingLogWriter(filepaths.EnsureLogDir(), "plugin"))

	logger := sdklogging.NewLogger(&hclog.LoggerOptions{
		Output:     writer,
		TimeFn:     func() time.Time { return time.Now().UTC() },
		TimeFormat: "2006-01-02 15:04:05.000 UTC",
	})
	log.SetOutput(logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)
	return logger
}
