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
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/plugin_manager"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"
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

func testPluginManagerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "test-plugin-manager",
		Run:    runTestPluginManagerCmd,
		Hidden: true,
	}
	cmdconfig.OnCmd(cmd)

	return cmd
}
func runTestPluginManagerCmd(cmd *cobra.Command, args []string) {
	log.Printf("[WARN] runTestPluginManagerCmd")
	man, err := plugin_manager.GetPluginManager()
	if err != nil {
		log.Printf("[WARN] error getting plugin manager: %s", err)
		return
	}
	log.Printf("[WARN] got plugin manager")

	plugin, err := man.Get(&pb.GetRequest{Connection: "aws"})
	if err != nil {
		log.Printf("[WARN] error getting plugin manager: %s", err)
		return
	}
	log.Printf("[WARN] got plugin %v", plugin)

}

func runPluginManagerCmd(cmd *cobra.Command, args []string) {
	logger := createPluginManagerLog()

	// build config map
	steampipeConfig, err := steampipeconfig.LoadConnectionConfig()
	if err != nil {
		utils.ShowError(err)
		os.Exit(1)
	}
	configMap := make(map[string]*pb.ConnectionConfig)
	for k, v := range steampipeConfig.Connections {
		configMap[k] = &pb.ConnectionConfig{
			Plugin:          v.Plugin,
			PluginShortName: v.PluginShortName,
			Config:          v.Config,
		}
	}
	plugin_manager.NewPluginManager(configMap, logger).Serve()
}

func createPluginManagerLog() hclog.Logger {
	logName := fmt.Sprintf("plugin-%s.log", time.Now().Format("2006-01-02"))
	logPath := filepath.Join(constants.LogDir(), logName)
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("failed to open plugin manager log file: %s\n", err.Error())
		os.Exit(1)
	}
	logger := logging.NewLogger(&hclog.LoggerOptions{Output: f})
	log.SetOutput(logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)
	return logger
}
