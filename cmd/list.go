package cmd

import (
	"context"
	"fmt"
	"sort"

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
	parentCmd *cobra.Command
}

func getListSubCmd(opts listSubCmdOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "list",
		TraverseChildren: true,
		Args:             cobra.NoArgs,
		Run:              getRunListSubCmd(opts),
		Short:            fmt.Sprintf("List all resources that can be executed with the '%s' command", opts.parentCmd.Name()),
		Long: fmt.Sprintf(`
List all resources that can be executed with the '%s' command.
`, opts.parentCmd.Name()),
	}

	cmdconfig.
		OnCmd(cmd)

	return cmd
}

// getRunListSubCmd generates a command handler based
// on the command that the runner is used for
func getRunListSubCmd(opts listSubCmdOptions) func(cmd *cobra.Command, args []string) {
	if opts.parentCmd == nil {
		// this should never happen
		panic("parent is required")
	}

	return func(cmd *cobra.Command, args []string) {
		workspacePath := viper.GetString(constants.ArgModLocation)

		w, err := workspace.Load(cmd.Context(), workspacePath)
		error_helpers.FailOnError(err)

		resourceTypesToDisplay := getResourceTypesToDisplay(cmd)
		resources, err := listResourcesInMod(cmd.Context(), w.Mod, resourceTypesToDisplay)
		error_helpers.FailOnErrorWithMessage(err, "could not list resources")

		sortResources(resources, w)
		headers, rows := getOutputDataTable(resources, w)

		display.ShowWrappedTable(headers, rows, &display.ShowWrappedTableOptions{
			AutoMerge:        false,
			HideEmptyColumns: true,
		})
	}

}

// listResourcesInMod walks through the resources in the given mod and
// uses the function to filter.
//
// if an error occurs, this function returns the list as has been generated till the error occured
// with the error
func listResourcesInMod(ctx context.Context, mod *modconfig.Mod, resourceTypes map[string]bool) ([]modconfig.ModTreeItem, error) {
	items := []modconfig.ModTreeItem{}
	err := mod.WalkResources(func(item modconfig.HclResource) (bool, error) {
		if ctx.Err() != nil {
			// break
			return false, ctx.Err()
		}

		// we need to 'cast' this since the GetParents is available only in the
		// ModTreeItem interface
		if cast, ok := item.(modconfig.ModTreeItem); ok {
			if resourceTypes[cast.BlockType()] && cast.GetParents()[0] == mod {
				items = append(items, cast)
			}
		}
		return true, nil
	})
	return items, err
}

func sortResources(items []modconfig.ModTreeItem, workspace *workspace.Workspace) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].GetUnqualifiedName() < items[j].GetUnqualifiedName()
	})
}

func getOutputDataTable(items []modconfig.ModTreeItem, workspace *workspace.Workspace) ([]string, [][]string) {
	rows := make([][]string, len(items))
	for idx, modItem := range items {
		rows[idx] = []string{modItem.GetUnqualifiedName(), modItem.GetTitle()}
	}
	return []string{"Name", "Title"}, rows
}

func getResourceTypesToDisplay(cmd *cobra.Command) map[string]bool {
	parent := cmd.Parent().Name()
	cmdToTypeMapping := map[string][]string{
		"check":     {"benchmark", "control"},
		"dashboard": {"dashboard", "benchmark"},
		"query":     {"query"},
	}
	resourceTypesToList, found := cmdToTypeMapping[parent]
	if !found {
		panic(fmt.Sprintf("could not find resource type lookup list for '%s'", parent))
	}
	// add resource types to a map for cheap lookup
	res := map[string]bool{}
	for _, t := range resourceTypesToList {
		res[t] = true
	}
	return res
}
