package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/turbot/steampipe/pkg/cmdconfig"
)

// TODO #kai can we just remove this
type listSubCmdOptions struct {
	parentCmd *cobra.Command
}

func getListSubCmd(opts listSubCmdOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "list",
		TraverseChildren: true,
		Args:             cobra.NoArgs,
		Run:              getRunListSubCmd(opts),
		Short:            fmt.Sprintf("List all resources that can be executed with the '%s' command", opts.parentCmd.Name()),
		Long:             fmt.Sprintf("List all resources that can be executed with the '%s' command", opts.parentCmd.Name()),
	}

	cmdconfig.
		OnCmd(cmd)

	return cmd
}

// getRunListSubCmd generates a command handler based on the parent command
func getRunListSubCmd(opts listSubCmdOptions) func(cmd *cobra.Command, args []string) {
	if opts.parentCmd == nil {
		// this should never happen
		panic("parent is required")
	}

	return func(cmd *cobra.Command, _ []string) {
		// TODO #v1 list query files? or deprecate list commena
		//ctx := cmd.Context()
		//
		//headers, rows := getOutputDataTable(modResources, depResources)
		//
		//display.ShowWrappedTable(headers, rows, &display.ShowWrappedTableOptions{
		//	AutoMerge:        false,
		//	HideEmptyColumns: true,
		//	Truncate:         true,
		//})
	}
}
