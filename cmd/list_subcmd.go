package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/cmdconfig"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/display"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/types"
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
		// we can panic here, since this will always come up during development
		panic("parent is required")
	}
	return func(cmd *cobra.Command, args []string) {
		parent := cmd.Parent().Name()
		cmdToTypeMapping := map[string][]string{
			"check": {"benchmark", "control"},
		}
		typ, found := cmdToTypeMapping[parent]
		if !found {
			typ = []string{cmd.Parent().Name()}
		}

		workspacePath := viper.GetString(constants.ArgModLocation)
		log.Println("[TRACE] workspace path:", workspacePath)

		wrkspc, err := workspace.Load(cmd.Context(), workspacePath)
		log.Println("[TRACE] workspace loaded:")
		error_helpers.FailOnError(err)

		// construct a Set with the resource types (Set uses a map under the hood)
		// that way, look ups are going to be cheaper
		// we need this optimization since a workspace can have
		// a huge number of resources and we need to iterate over all of them
		lookupResourceTypes := types.NewSet[string]()
		for _, t := range typ {
			lookupResourceTypes.Add(t)
		}

		items := []modconfig.ModTreeItem{}

		log.Println("[TRACE] listing:", lookupResourceTypes)
		wrkspc.Mod.WalkResources(func(item modconfig.HclResource) (bool, error) {
			if lookupResourceTypes.Has(item.BlockType()) {
				if cast, ok := item.(modconfig.ModTreeItem); ok {
					items = append(items, cast)
				}
			}
			return true, nil
		})

		rows := [][]string{}
		for _, modItem := range items {
			rows = append(rows, []string{modItem.GetUnqualifiedName(), modItem.GetTitle()})
		}
		display.ShowWrappedTable([]string{"name", "title"}, rows, &display.ShowWrappedTableOptions{
			AutoMerge:        false,
			HideEmptyColumns: true,
		})
	}

}
