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

	return func(cmd *cobra.Command, args []string) {
		workspacePath := viper.GetString(constants.ArgModLocation)

		w, err := workspace.Load(cmd.Context(), workspacePath)
		error_helpers.FailOnError(err)

		modResources, depResources, err := listResourcesInMod(cmd.Context(), w.Mod, cmd)
		error_helpers.FailOnErrorWithMessage(err, "could not list resources")
		if len(modResources)+len(depResources) == 0 {
			fmt.Println("No resources available to execute.")
		}

		sortResources(modResources)
		sortResources(depResources)
		headers, rows := getOutputDataTable(modResources, depResources)

		display.ShowWrappedTable(headers, rows, &display.ShowWrappedTableOptions{
			AutoMerge:        false,
			HideEmptyColumns: true,
			Truncate:         true,
		})
	}
}

func listResourcesInMod(ctx context.Context, mod *modconfig.Mod, cmd *cobra.Command) (modResources, depResources []modconfig.ModTreeItem, err error) {
	resourceTypesToDisplay := getResourceTypesToDisplay(cmd)

	err = mod.WalkResources(func(item modconfig.HclResource) (bool, error) {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}

		// if we are not showing this resource type, return
		if !resourceTypesToDisplay[item.BlockType()] {
			return true, nil
		}

		m := item.(modconfig.ModTreeItem)

		itemMod := m.GetMod()
		if m.GetParents()[0] == itemMod {

			// add to the appropriate array
			if itemMod.Name() == mod.Name() {
				modResources = append(modResources, m)
			} else {
				depResources = append(depResources, m)
			}
		}
		return true, nil
	})
	return modResources, depResources, err
}

func sortResources(items []modconfig.ModTreeItem) {
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Name() < items[j].Name()
	})
}

func getOutputDataTable(modResources, depResources []modconfig.ModTreeItem) ([]string, [][]string) {
	rows := make([][]string, len(modResources)+len(depResources))
	for i, modItem := range modResources {
		rows[i] = []string{modItem.GetUnqualifiedName(), modItem.GetTitle()}
	}
	offset := len(modResources)
	for i, modItem := range depResources {
		// use fully qualified name for dependency resources
		rows[i+offset] = []string{modItem.Name(), modItem.GetTitle()}
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
	resourceTypesToDisplay, found := cmdToTypeMapping[parent]
	if !found {
		panic(fmt.Sprintf("could not find resource type lookup list for '%s'", parent))
	}
	// add resource types to a map for cheap lookup
	res := map[string]bool{}
	for _, t := range resourceTypesToDisplay {
		res[t] = true
	}
	return res
}
