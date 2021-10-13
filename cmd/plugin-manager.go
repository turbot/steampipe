package cmd

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/plugin_manager"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/utils"
)

// pluginManagerCmd :: represents the pluginManager command
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
	//if viper.GetBool("spawn") {
	//	spawnPluginManagerCommand()
	//	return
	//}

	steampipeConfig, err := steampipeconfig.LoadConnectionConfig()
	if err != nil {
		utils.ShowError(err)
		return
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
	plugin_manager.NewPluginManager(configMap).Serve()
}

func spawnPluginManagerCommand() {
	// create command which will run steampipe in plugin-manager mode
	pluginManagerCmd := exec.Command("steampipe", "plugin-manager", "spawn=false")
	pluginManagerCmd.Start()

	// wait for someone to kill us
	for {
	}

}

// pluginManagerCmd :: represents the pluginManager command
func startPluginManagerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "start-plugin-manager",
		Run:    runStartPluginManagerCmd,
		Hidden: true,
	}

	cmdconfig.OnCmd(cmd).AddStringFlag(constants.ArgSeparator, "", ",", "Separator string for csv output")

	return cmd
}

func runStartPluginManagerCmd(cmd *cobra.Command, args []string) {
	// we want to see the plugin manager log
	log.SetOutput(os.Stdout)

	// create command which will run steampipe in plugin-manager mode
	pluginManagerCmd := exec.Command("steampipe", "plugin-manager")
	pluginManagerCmd.Stdout = os.Stdout
	pluginManagerCmd.Start()

	//// launch the plugin manager the plugin process.
	//client := plugin.NewClient(&plugin.ClientConfig{
	//	HandshakeConfig: pluginshared.Handshake,
	//	Plugins:         pluginshared.PluginMap,
	//	Cmd:             pluginManagerCmd,
	//	AllowedProtocols: []plugin.Protocol{
	//		plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	//})
	//_, err := client.Start()
	//utils.FailOnError(err)
	//
	//// create a plugin manager state
	//state := plugin_manager.NewPluginManagerState(client.ReattachConfig())
	//
	//// now save the state
	//err = state.Save()
	//utils.FailOnError(err)

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
