package controldisplay

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/cloud"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/control/controlexecute"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardexecute"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func executionTreeToSnapshot(e *controlexecute.ExecutionTree) (*dashboardtypes.SteampipeSnapshot, error) {
	var dashboardNode modconfig.DashboardLeafNode
	var panels map[string]dashboardtypes.SnapshotPanel
	var checkRun *dashboardexecute.CheckRun
	var nodeType string

	// get root benchmark/control
	switch root := e.Root.Children[0].(type) {
	case *controlexecute.ResultGroup:
		var ok bool
		dashboardNode, ok = root.GroupItem.(modconfig.DashboardLeafNode)
		if !ok {
			return nil, fmt.Errorf("invalid node found in control execution tree - cannot cast '%s' to a DashboardLeafNode", root.GroupItem.Name())
		}
		nodeType = "benchmark"
	case *controlexecute.ControlRun:
		dashboardNode = root.Control
		nodeType = "control"
	}

	// create a check run to wrap the execution tree
	checkRun = &dashboardexecute.CheckRun{
		Root:          e.Root.Children[0],
		Name:          dashboardNode.Name(),
		DashboardNode: dashboardNode,
		NodeType:      nodeType,
		DashboardName: dashboardNode.Name(),
		Title:         dashboardNode.GetTitle(),
		Description:   dashboardNode.GetDescription(),
	}

	// populate the panels
	panels = checkRun.BuildSnapshotPanels(make(map[string]dashboardtypes.SnapshotPanel))

	// create the snapshot
	res := &dashboardtypes.SteampipeSnapshot{
		SchemaVersion: fmt.Sprintf("%d", dashboardtypes.SteampipeSnapshotSchemaVersion),
		Panels:        panels,
		Layout:        checkRun.Root.AsTreeNode(),
		Inputs:        map[string]interface{}{},
		Variables:     dashboardexecute.GetReferencedVariables(checkRun, e.Workspace),
		SearchPath:    e.SearchPath,
		StartTime:     e.StartTime,
		EndTime:       e.EndTime,
		Title:         dashboardNode.GetTitle(),
		FileNameRoot:  dashboardNode.Name(),
	}
	return res, nil
}

func PublishSnapshot(ctx context.Context, e *controlexecute.ExecutionTree, shouldShare bool) error {
	snapshot, err := executionTreeToSnapshot(e)
	if err != nil {
		return err
	}

	message, err := cloud.PublishSnapshot(ctx, snapshot, shouldShare)
	if err != nil {
		return err
	}
	if viper.GetBool(constants.ArgProgress) {
		fmt.Println(message)
	}
	return nil

}
