package cmdconfig

import "github.com/spf13/cobra"

func IsServiceStopCmd(cmd *cobra.Command) bool {
	return cmd.Parent() != nil && cmd.Parent().Name() == "service" && cmd.Name() == "stop"
}
func IsCompletionCmd(cmd *cobra.Command) bool {
	return cmd.Name() == "completion"
}
func IsPluginManagerCmd(cmd *cobra.Command) bool {
	return cmd.Name() == "plugin-manager"
}
func IsPluginUpdateCmd(cmd *cobra.Command) bool {
	return cmd.Name() == "update" && cmd.Parent() != nil && cmd.Parent().Name() == "plugin"
}
func IsBatchQueryCmd(cmd *cobra.Command, cmdArgs []string) bool {
	return cmd.Name() == "query" && len(cmdArgs) > 0
}
