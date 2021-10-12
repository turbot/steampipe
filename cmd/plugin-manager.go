package cmd

import (
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
	cmdconfig.OnCmd(cmd).AddStringFlag(constants.ArgSeparator, "", ",", "Separator string for csv output")

	return cmd
}

func runPluginManagerCmd(cmd *cobra.Command, args []string) {
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
