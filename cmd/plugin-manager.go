package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/utils"

	"github.com/hashicorp/hcl/v2"

	"github.com/hashicorp/go-plugin"
	pluginshared "github.com/turbot/steampipe-plugin-sdk/grpc/shared"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/plugin_manager"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"
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

func testPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "test-plugin",
		Run:    runTestPluginCmd,
		Hidden: true,
	}
	cmdconfig.OnCmd(cmd)
	return cmd
}

func runTestPluginCmd(cmd *cobra.Command, args []string) {
	log.Printf("[WARN] runTestPluginCmd")
	err := CreateConnectionPlugin(&modconfig.Connection{
		Name:            "aws",
		PluginShortName: "aws",
		Plugin:          "hub.steampipe.io/plugins/turbot/aws@latest",
		Config:          "",
		Options:         nil,
		DeclRange:       hcl.Range{},
	})
	if err != nil {
		log.Printf("[WARN] error getting plugin manager: %s", err)
	}

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

func CreateConnectionPlugin(connection *modconfig.Connection) error {

	//plugin := "hub.steampipe.io/plugins/turbot/aws@latest"

	remoteSchema := connection.Plugin
	connectionName := connection.Name

	pluginPath, err := plugin_manager.GetPluginPath(remoteSchema, connectionName)
	if err != nil {
		return err
	}

	log.Printf("[WARN] got plugin path %s", pluginPath)

	// launch the plugin process.
	// create the plugin map
	pluginMap := map[string]plugin.Plugin{
		remoteSchema: &pluginshared.WrapperPlugin{},
	}
	loggOpts := &hclog.LoggerOptions{Name: "plugin"}
	logger := logging.NewLogger(loggOpts)

	cmd := exec.Command(pluginPath)
	// pass env to command
	cmd.Env = os.Environ()
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginMap,
		// this failed when running from extension
		//Cmd:              exec.Command("sh", "-c", pluginPath),
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})

	log.Printf("[WARN] got client")
	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return err
	}

	log.Printf("[WARN] got rpc client")
	// Request the plugin
	raw, err := rpcClient.Dispense(remoteSchema)
	if err != nil {
		return err
	}
	log.Printf("[WARN] got raw client")
	// We should have a stub plugin now
	p := raw.(pluginshared.WrapperPluginClient)

	log.Printf("[WARN] got stub %p", p)

	return nil
}
