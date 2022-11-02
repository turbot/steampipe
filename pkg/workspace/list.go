package workspace

import (
	"context"
	"log"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/display"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func (w *Workspace) ListWorkspaceResources(ctx context.Context, resourceTypes ...string) ([]modconfig.HclResource, error) {
	listed := []modconfig.HclResource{}

	log.Println("[TRACE] listing:", resourceTypes)
	w.Mod.WalkResources(func(item modconfig.HclResource) (bool, error) {
		if helpers.StringSliceContains(resourceTypes, item.BlockType()) {
			listed = append(listed, item)
		}
		return true, nil
	})
	return listed, nil
}

func ListResources(ctx context.Context, types ...string) {
	workspacePath := viper.GetString(constants.ArgModLocation)
	log.Println("[TRACE] workspace path:", workspacePath)

	w, err := Load(ctx, workspacePath)
	log.Println("[TRACE] workspace loaded:")
	error_helpers.FailOnError(err)

	queries, err := w.ListWorkspaceResources(ctx, types...)
	log.Println("[TRACE] workspace listed:")
	log.Println("[TRACE] list len:", len(queries))
	error_helpers.FailOnError(err)

	rows := [][]string{}

	for _, q := range queries {
		if modItem, ok := q.(modconfig.ModTreeItem); ok {
			rows = append(rows, []string{modItem.GetUnqualifiedName(), modItem.GetTitle()})
		}
	}

	display.ShowWrappedTable([]string{"name", "title"}, rows, false)
}
