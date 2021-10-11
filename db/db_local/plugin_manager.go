package db_local

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"

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
		Cmd:             exec.Command("sh", "-c", "steampipe plugin-manager"),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	})
	if _, err := client.Start(); err != nil {
		return err
	}

	reattach := client.ReattachConfig()

	// now save the reattach config
	return savePluginManagerReattachConfig(reattach)
	//// now attach using this reaattach config
	//newClient := plugin.NewClient(&plugin.ClientConfig{
	//	HandshakeConfig: pluginshared.Handshake,
	//	Plugins:         pluginshared.PluginMap,
	//	Reattach:        reattach,
	//	AllowedProtocols: []plugin.Protocol{
	//		plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	//})
	//
	//// Connect via RPC
	//rpcClient, err := newClient.Client()
	//if err != nil {
	//	fmt.Println("Error:", err.Error())
	//	os.Exit(1)
	//}
	//
	//// Request the plugin
	//raw, err := rpcClient.Dispense(pluginshared.PluginName)
	//if err != nil {
	//	fmt.Println("Error:", err.Error())
	//	os.Exit(1)
	//}
	//
	//// We should have a KV store now! This feels like a normal interface
	//// implementation but is in fact over an RPC connection.
	//pluginManager := raw.(pluginshared.PluginManager)
	//
	//resp, err := pluginManager.GetPlugin(&pb.GetPluginRequest{})
	//fmt.Println(resp)
	//return err
}

func savePluginManagerReattachConfig(reattach *plugin.ReattachConfig) error {
	content, err := json.Marshal(inreattachfo)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(runningInfoFilePath(), content, 0644)

}
