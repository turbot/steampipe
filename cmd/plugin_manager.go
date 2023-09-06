package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/logging"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/connection"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/pluginmanager_service"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
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
	logger := createPluginManagerLog()

	log.Printf("[INFO] starting plugin manager")
	// build config map
	steampipeConfig, errorsAndWarnings := steampipeconfig.LoadConnectionConfig()
	if errorsAndWarnings.GetError() != nil {
		log.Printf("[WARN] failed to load connection config: %v", errorsAndWarnings.GetError())
		os.Exit(1)
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

	configMap := connection.NewConnectionConfigMap(steampipeConfig.Connections)
	log.Printf("[TRACE] loaded config map: %s", strings.Join(steampipeConfig.ConnectionNames(), ","))

	pluginManager, err := pluginmanager_service.NewPluginManager(cmd.Context(), configMap, steampipeConfig.Plugins, logger)
	if err != nil {
		log.Printf("[WARN] failed to create plugin manager: %s", err.Error())
		os.Exit(1)
	}

	if shouldRunConnectionWatcher() {
		log.Printf("[INFO] starting connection watcher")
		connectionWatcher, err := connection.NewConnectionWatcher(pluginManager)
		if err != nil {
			log.Printf("[WARN] failed to create connection watcher: %s", err.Error())
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

	// we use this logger to log from the plugin processes
	// the plugin processes uses the `EscapeNewlineWriter` to map the '\n' byte to "\n" string literal
	// this is to allow the plugin to send multiline log messages as a single log line.
	//
	// here we apply the reverse mapping to get back the original message
	writer := logging.NewUnescapeNewlineWriter(f)

	logger := logging.NewLogger(&hclog.LoggerOptions{
		Output:     writer,
		TimeFn:     func() time.Time { return time.Now().UTC() },
		TimeFormat: "2006-01-02 15:04:05.000 UTC",
	})
	log.SetOutput(logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)
	return logger
}
