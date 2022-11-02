package cmd

import (
	"github.com/spf13/cobra"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/workspace"
)

func listSubCmd(shortDescription string, longDecsription string) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "list",
		TraverseChildren: true,
		Args:             cobra.NoArgs,
		Run:              runListSubCmd,
		Short:            shortDescription,
		Long:             longDecsription,
	}

	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag(constants.ArgAll, "", false, "List ALL items that can be run")

	return cmd
}

func runListSubCmd(cmd *cobra.Command, args []string) {
	parent := cmd.Parent().Name()
	resMap := map[string][]string{
		"check": {"benchmark", "control"},
	}
	typ, found := resMap[parent]
	if !found {
		typ = []string{cmd.Parent().Name()}
	}
	workspace.ListResources(cmd.Context(), typ...)
}
