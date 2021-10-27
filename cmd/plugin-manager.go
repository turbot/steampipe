package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/spf13/cobra"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	sdkshared "github.com/turbot/steampipe-plugin-sdk/grpc/shared"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmdconfig"
	"github.com/turbot/steampipe/plugin_manager"
	pb "github.com/turbot/steampipe/plugin_manager/grpc/proto"
	pluginshared "github.com/turbot/steampipe/plugin_manager/grpc/shared"
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
	startPluginManager()
	//if viper.GetBool("spawn") {
	//	spawnPluginManager()
	//	return
	//} else {
	//startPluginManager()
	//}
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
		fmt.Println("NOOOO")
		os.Exit(1)
	}
	logger := logging.NewLogger(&hclog.LoggerOptions{Output: f})
	log.SetOutput(f)
	log.Printf("[WARN] FOOOOOO")

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
	plugin_manager.NewPluginManager(configMap, logger).Serve()
}

func startPluginManager2() {
	logfile := "/tmp/plugin_manager.log"
	f, _ := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	logger := logging.NewLogger(&hclog.LoggerOptions{Output: f})
	log.SetOutput(f)

	log.Printf("[WARN] plugin manager logging")

	// get connection config

	pluginName := "hub.steampipe.io/plugins/turbot/chaos@latest"
	pluginPath, _ := plugin_manager.GetPluginPath(pluginName, "chaos")

	// create the plugin map
	pluginMap := map[string]plugin.Plugin{
		pluginName: &sdkshared.WrapperPlugin{},
	}
	//loggOpts := &hclog.LoggerOptions{Name: "plugin"}
	//logger := logging.NewLogger(loggOpts)

	cmd := exec.Command(pluginPath)
	// pass env to command
	cmd.Env = os.Environ()
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  sdkshared.Handshake,
		Plugins:          pluginMap,
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})

	defer client.Kill()

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(pluginName)
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// We should have a KV store now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	c := raw.(sdkshared.WrapperPluginClient)

	c.GetSchema(&proto.GetSchemaRequest{})
	c.GetSchema(&proto.GetSchemaRequest{})
}

func startPluginManager3() {
	logfile := "/tmp/plugin_manager.log"
	f, _ := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	logger := logging.NewLogger(&hclog.LoggerOptions{Output: f})
	log.SetOutput(f)

	log.Printf("[WARN] plugin manager logging")

	// get connection config

	log.Printf("[WARN] startPlugin ********************\n")

	pluginName := "hub.steampipe.io/plugins/turbot/chaos@latest"
	pluginPath, _ := plugin_manager.GetPluginPath(pluginName, "chaos")

	// create the plugin map
	pluginMap := map[string]plugin.Plugin{
		pluginName: &sdkshared.WrapperPlugin{},
	}

	cmd := exec.Command(pluginPath)
	// pass env to command
	cmd.Env = os.Environ()
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  sdkshared.Handshake,
		Plugins:          pluginMap,
		Cmd:              cmd,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Logger:           logger,
	})

	client.Start()
	// create grpc client
	client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  pluginshared.Handshake,
		Plugins:          pluginMap,
		Reattach:         client.ReattachConfig(),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		//Logger:           logger,
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(pluginName)
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// We should have a KV store now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	c := raw.(sdkshared.WrapperPluginClient)

	c.GetSchema(&proto.GetSchemaRequest{})
	c.GetSchema(&proto.GetSchemaRequest{})
}

func startPluginManager4() {
	logfile := "/tmp/plugin_manager.log"
	f, _ := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	logger := logging.NewLogger(&hclog.LoggerOptions{Output: f})
	log.SetOutput(f)

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
	p := plugin_manager.NewPluginManager(configMap, logger)

	r, _ := p.Get(&pb.GetRequest{
		Connection: "chaos",
	})

	pluginName := "hub.steampipe.io/plugins/turbot/chaos@latest"

	// create the plugin map
	pluginMap := map[string]plugin.Plugin{
		pluginName: &sdkshared.WrapperPlugin{},
	}

	// create grpc client
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  pluginshared.Handshake,
		Plugins:          pluginMap,
		Reattach:         r.Reattach.Convert(),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		//Logger:           logger,
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(pluginName)
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// We should have a KV store now! This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	c := raw.(sdkshared.WrapperPluginClient)

	c.GetSchema(&proto.GetSchemaRequest{})
	c.GetSchema(&proto.GetSchemaRequest{})
}

func logOutput() func() {
	logfile := "/tmp/plugin_manager.log"
	// open file read/write | create if not exist | clear file at open if exists
	f, _ := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)

	// save existing stdout | MultiWriter writes to saved stdout and file
	out := os.Stdout
	mw := io.MultiWriter(out, f)

	// get pipe reader and writer | writes to pipe writer come out pipe reader
	r, w, _ := os.Pipe()

	// replace stdout,stderr with pipe writer | all writes to stdout, stderr will go through pipe instead (fmt.print, log)
	os.Stdout = w
	os.Stderr = w

	// writes with log.Print should also write to mw
	log.SetOutput(mw)

	//create channel to control exit | will block until all copies are finished
	exit := make(chan bool)

	go func() {
		// copy all reads from pipe to multiwriter, which writes to stdout and file
		_, _ = io.Copy(mw, r)
		// when r or w is closed copy will finish and true will be sent to channel
		exit <- true
	}()

	// function to be deferred in main until program exits
	return func() {
		// close writer then block on exit channel | this will let mw finish writing before the program exits
		_ = w.Close()
		<-exit
		// close file after all writes have finished
		_ = f.Close()
	}

}

// newHCLogger returns a new hclog.Logger instance with the given name
func newHCLogger(name string) hclog.Logger {
	logOutput := io.Writer(os.Stderr)
	f, err := os.OpenFile("/tmp/plugin_manager.log", syscall.O_CREAT|syscall.O_RDWR|syscall.O_APPEND, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
	} else {
		logOutput = f
	}

	return hclog.NewInterceptLogger(&hclog.LoggerOptions{
		Name:              name,
		Level:             hclog.Trace,
		Output:            logOutput,
		IndependentLevels: true,
	})
}
