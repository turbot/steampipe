package cmd

import (
	"context"
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
	shortDescription string
	longDescription  string
	allDescription   string
	parentCmd        *cobra.Command
}

func setListSubCmdOptionsDefaults(opts *listSubCmdOptions) {
	// these should never happen
	if opts == nil {
		panic("Options MUST be set")
	}
	if opts.parentCmd == nil {
		panic("The Parent CMD MUST be set")
	}

	if len(opts.shortDescription) == 0 {
		opts.shortDescription = "List all resources that can be executed"
	}
	if len(opts.longDescription) == 0 {
		opts.longDescription = "long description placeholder"
	}
	if len(opts.allDescription) == 0 {
		opts.allDescription = "all flag description placeholder"
	}
}

func getListSubCmd(opts listSubCmdOptions) *cobra.Command {

	setListSubCmdOptionsDefaults(&opts)

	cmd := &cobra.Command{
		Use:              "list",
		TraverseChildren: true,
		Args:             cobra.NoArgs,
		Run:              getRunListSubCmdRun(opts),
		Short:            opts.shortDescription,
		Long:             opts.longDescription,
	}

	cmdconfig.
		OnCmd(cmd).
		AddBoolFlag(constants.ArgAll, "", false, opts.allDescription)

	return cmd
}

// getRunListSubCmdRun generates a command handler based
// on the command that the runner is used for
func getRunListSubCmdRun(opts listSubCmdOptions) func(cmd *cobra.Command, args []string) {
	if opts.parentCmd == nil {
		// this should never happen
		panic("parent is required")
	}

	return func(cmd *cobra.Command, args []string) {
		workspacePath := viper.GetString(constants.ArgModLocation)

		w, err := workspace.Load(cmd.Context(), workspacePath)
		error_helpers.FailOnError(err)

		resourceTypesToDisplay := getResourceTypesToDisplay(cmd)
		resources, err := listResourcesInMod(cmd.Context(), w.Mod, func(item modconfig.ModTreeItem) bool {
			return resourceTypesToDisplay[item.BlockType()]
		})
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
func listResourcesInMod(ctx context.Context, mod *modconfig.Mod, filter func(modconfig.ModTreeItem) bool) ([]modconfig.ModTreeItem, error) {
	items := []modconfig.ModTreeItem{}
	err := mod.WalkResources(func(item modconfig.HclResource) (bool, error) {
		if ctx.Err() != nil {
			// break
			return false, ctx.Err()
		}
		if cast, ok := item.(modconfig.ModTreeItem); ok {
			if filter(cast) {
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
	return []string{"name", "title"}, rows
}

func getResourceTypesToDisplay(cmd *cobra.Command) map[string]bool {
	parent := cmd.Parent().Name()
	cmdToTypeMapping := map[string][]string{
		"check":     {"benchmark", "control"},
		"dashboard": {"dashboard", "benchmark"},
	}
	xtraTypesForAll := map[string][]string{}

	resourceTypesToList, found := cmdToTypeMapping[parent]
	if !found {
		resourceTypesToList = []string{cmd.Parent().Name()}
	}
	// add resource types to a map for cheap lookup
	lookupResourceTypes := map[string]bool{}
	for _, t := range resourceTypesToList {
		lookupResourceTypes[t] = true
	}

	// if the '--all' flag is set
	if viper.GetBool(constants.ArgAll) {
		xtraTypesToList, found := xtraTypesForAll[parent]
		if !found {
			xtraTypesToList = []string{cmd.Parent().Name()}
		}
		for _, t := range xtraTypesToList {
			lookupResourceTypes[t] = true
		}
	}
	return lookupResourceTypes
}
