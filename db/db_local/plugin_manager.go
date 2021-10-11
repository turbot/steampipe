package db_local

import (
	"fmt"
	"os"
	"os/exec"

	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"

	"github.com/hashicorp/go-plugin"
	pluginshared "github.com/turbot/steampipe/plugin_manager/grpc/shared"
)

func StartPluginManager() error {
	// We don't want to see the plugin logs.
	//log.SetOutput(ioutil.Discard)

	// launch the plugin manager the plugin process.
	// TODO pass config path or set connection config
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginshared.PluginMap,
		Cmd:             exec.Command("sh", "-c", pluginshared.PluginName),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	})
	client.Start()

	reattach := client.ReattachConfig()

	// now attach using this reaattach config
	plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: pluginshared.Handshake,
		Plugins:         pluginshared.PluginMap,
		Reattach:        reattach,
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(pluginshared.PluginName)
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// We should have a KV store now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	pluginManager := raw.(pluginshared.PluginManager)

	resp, err := pluginManager.GetPlugin(&pb.GetPluginRequest{})
	fmt.Println(resp)
	return err
}
