package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/turbot/steampipe-plugin-sdk/logging"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-plugin/examples/grpc/shared"
)

func main() {
	// We don't want to see the plugin logs.
	//log.SetOutput(ioutil.Discard)
	logfile := "/tmp/grpc.log"
	// open file read/write | create if not exist | clear file at open if exists
	f, _ := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)

	logger := logging.NewLogger(&hclog.LoggerOptions{Output: f})

	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         shared.PluginMap,
		Cmd:             exec.Command("sh", "-c", "/Users/kai/Dev/github/turbot/steampipe/tests/grpc/kv-go-grpc"),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
		Logger: logger,
	})
	defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("kv_grpc")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// We should have a KV store now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	kv := raw.(shared.KV)

	kv.Get("")

	os.Exit(0)
}
