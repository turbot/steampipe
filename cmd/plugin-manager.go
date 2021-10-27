package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmdconfig"
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
	cmdconfig.OnCmd(cmd).
		AddBoolFlag("spawn", "", false, "")

	return cmd
}

func runPluginManagerCmd(cmd *cobra.Command, args []string) {
	if viper.GetBool("spawn") {
		spawnPluginManager()
	} else {
		startPluginManager()
	}
}

func spawnPluginManager() {
	// create command which will run steampipe in plugin-manager mode
	pluginManagerCmd := exec.Command("steampipe", "plugin-manager")
	pluginManagerCmd.Stdout = os.Stdout
	pluginManagerCmd.Start()

	// wait to be killed
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	<-sigchan

	// kill our child
	// NOTE we will not do this if kill -9 is run
	pluginManagerCmd.Process.Kill()
}

func startPluginManager() {
	// TODO get install dir (or ensure this is running from install dir)
	logfile := "/tmp/plugin_manager.log"
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("failed to open plugin manager log file: %s\n", err.Error())
		os.Exit(1)
	}
	logger := logging.NewLogger(&hclog.LoggerOptions{Output: f})
	log.SetOutput(f)
	//log.Println("[WARN] cwd ", process.c)

	steampipeConfig, err := steampipeconfig.LoadConnectionConfig()
	if err != nil {
		utils.ShowError(err)
		os.Exit(1)
	}

	// build config map
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
