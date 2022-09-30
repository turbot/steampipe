package snapshot

import (
	"database/sql"
	"fmt"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

func QueryResultToSnapshot(resultRows [][]interface{}, colTypes []*sql.ColumnType, query modconfig.HclResource) (*dashboardtypes.SteampipeSnapshot, error) {

	//executionTree, err := NewDashboardExecutionTreeForQueryResultSnapshot(resultRows, colTypes, query)
	//if err != nil {
	//	return nil, err
	//}
	//// populate the panels
	//var panels = executionTree.BuildSnapshotPanels()
	//// create the snapshot
	//res := &dashboardtypes.SteampipeSnapshot{
	//	SchemaVersion: fmt.Sprintf("%d", dashboardtypes.SteampipeSnapshotSchemaVersion),
	//	Panels:        panels,
	//	Layout:        executionTree.Root.AsTreeNode(),
	//	Inputs:        map[string]interface{}{},
	//	//Variables:     dashboardexecute.GetReferencedVariables(executionTree.Root, e.Workspace),
	//	//SearchPath:    e.SearchPath,
	//	//StartTime:     e.StartTime,
	//	//EndTime:       e.EndTime,
	//}
	//

	res := &dashboardtypes.SteampipeSnapshot{
		SchemaVersion: fmt.Sprintf("%d", dashboardtypes.SteampipeSnapshotSchemaVersion),
		//Panels:        panels,
		//Layout:        checkRun.Root.AsTreeNode(),
		Inputs: map[string]interface{}{},
		//Variables:     dashboardexecute.GetReferencedVariables(checkRun, e.Workspace),
		//SearchPath:    e.SearchPath,
		//StartTime:     e.StartTime,
		//EndTime:       e.EndTime,
	}
	return res, nil
}

//
//// NewDashboardExecutionTreeForQueryResultSnapshot creates a shell DashboardExecutionTree that will be used for an
//// existing query result just to create a snapshot
//func NewDashboardExecutionTreeForQueryResultSnapshot(resultRows [][]interface{}, columns []*sql.ColumnType, query modconfig.HclResource) (*DashboardExecutionTree, error) {
//	dashboard, err := modconfig.NewQueryDashboard(query)
//	if err != nil {
//		return nil, err
//	}
//	// now populate the DashboardExecutionTree
//	executionTree := &DashboardExecutionTree{
//		dashboardName: dashboard.Name(),
//		runs:          make(map[string]dashboardtypes.DashboardNodeRun),
//	}
//
//	// now populate the DashboardExecutionTree
//	run, err := NewDashboardRun(dashboard, nil, executionTree)
//	if err != nil {
//		return nil, err
//	}
//	executionTree.Root = run
//	return executionTree, nil
//}
//
//type SnapshotExecutionTree struct {
//}
