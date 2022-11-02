package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/display"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/workspace"
)

type listSubCmdOptions struct {
	shortDescription string
	longDescription  string
	allDescription   string
	parent           *cobra.Command
}

func listSubCmd(opts listSubCmdOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "list",
		TraverseChildren: true,
		Args:             cobra.NoArgs,
		Run:              listSubCmdRunner(opts),
		Short:            opts.shortDescription,
		Long:             opts.longDescription,
	}

	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag(constants.ArgAll, "", false, opts.allDescription)

	return cmd
}

func listSubCmdRunner(opts listSubCmdOptions) func(cmd *cobra.Command, args []string) {
	if opts.parent == nil {
		// this should never happen
		panic("parent is required")
	}

	return func(cmd *cobra.Command, args []string) {
		workspacePath := viper.GetString(constants.ArgModLocation)

		w, err := workspace.Load(cmd.Context(), workspacePath)
		error_helpers.FailOnError(err)

		items := []modconfig.ModTreeItem{}
		resourceTypesToDisplay := getResourceTypesToDisplay(cmd)
		w.Mod.WalkResources(func(item modconfig.HclResource) (bool, error) {
			if _, found := resourceTypesToDisplay[item.BlockType()]; found {
				if cast, ok := item.(modconfig.ModTreeItem); ok {
					items = append(items, cast)
				}
			}
			return true, nil
		})

		rows := make([][]string, len(items))

		for idx, modItem := range items {
			rows[idx] = []string{modItem.GetUnqualifiedName(), modItem.GetTitle()}
		}
		display.ShowWrappedTable([]string{"name", "title"}, rows, &display.ShowWrappedTableOptions{
			AutoMerge:        false,
			HideEmptyColumns: true,
		})
	}

}

func getResourceTypesToDisplay(cmd *cobra.Command) map[string]struct{} {
	parent := cmd.Parent().Name()
	cmdToTypeMapping := map[string][]string{
		"check": {"benchmark", "control"},
	}
	resourceTypesToList, found := cmdToTypeMapping[parent]
	if !found {
		resourceTypesToList = []string{cmd.Parent().Name()}
	}
	// add resource types to a map for cheap lookup
	lookupResourceTypes := map[string]struct{}{}
	for _, t := range resourceTypesToList {
		lookupResourceTypes[t] = struct{}{}
	}
	return lookupResourceTypes
}
