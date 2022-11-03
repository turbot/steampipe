package cmd

import (
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
		opts.shortDescription = ""
	}
	if len(opts.longDescription) == 0 {
		opts.longDescription = ""
	}
	if len(opts.allDescription) == 0 {
		opts.allDescription = ""
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

		items := []modconfig.ModTreeItem{}
		resourceTypesToDisplay := getResourceTypesToDisplay(cmd)
		w.Mod.WalkResources(func(item modconfig.HclResource) (bool, error) {
			if found := resourceTypesToDisplay[item.BlockType()]; found {
				if cast, ok := item.(modconfig.ModTreeItem); ok {
					items = append(items, cast)
				}
			}
			return true, nil
		})

		sortItems(items, w)
		headers, rows := getOutputDataTable(items, w)

		display.ShowWrappedTable(headers, rows, &display.ShowWrappedTableOptions{
			AutoMerge:        false,
			HideEmptyColumns: true,
		})
	}

}

func sortItems(items []modconfig.ModTreeItem, workspace *workspace.Workspace) {
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
	resourceTypesToList, found := cmdToTypeMapping[parent]
	if !found {
		resourceTypesToList = []string{cmd.Parent().Name()}
	}
	// add resource types to a map for cheap lookup
	lookupResourceTypes := map[string]bool{}
	for _, t := range resourceTypesToList {
		lookupResourceTypes[t] = true
	}
	return lookupResourceTypes
}
