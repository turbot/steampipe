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
	var nodeType string
	var sourceDefinition string
	var dashboardNode modconfig.DashboardLeafNode

	// get root benchmark/control
	switch rootGroup := e.Root.Children[0].(type) {
	case *controlexecute.ResultGroup:
		sourceDefinition = rootGroup.GroupItem.(modconfig.ResourceWithMetadata).GetMetadata().SourceDefinition
		dashboardNode = rootGroup.GroupItem.(modconfig.DashboardLeafNode)
		nodeType = modconfig.BlockTypeBenchmark

	case *controlexecute.ControlRun:
		sourceDefinition = rootGroup.Control.GetMetadata().SourceDefinition
		dashboardNode = rootGroup.Control
		nodeType = modconfig.BlockTypeControl
	}

	// create a check run to wrap the execution tree
	checkRun := &dashboardexecute.CheckRun{
		Name:             dashboardNode.Name(),
		Title:            dashboardNode.GetTitle(),
		Description:      dashboardNode.GetDescription(),
		Documentation:    dashboardNode.GetDocumentation(),
		Display:          dashboardNode.GetDisplay(),
		Type:             dashboardNode.GetType(),
		Tags:             dashboardNode.GetTags(),
		DashboardName:    dashboardNode.GetUnqualifiedName(),
		SessionId:        "steampipe check",
		SourceDefinition: sourceDefinition,
		NodeType:         nodeType,
		DashboardNode:    dashboardNode,
		Summary:          e.Root.Summary,
		Root:             e.Root.Children[0],
	}

	// populate the panels
	var panels = checkRun.BuildSnapshotPanels(make(map[string]dashboardtypes.SnapshotPanel))

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
