package dashboardserver

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
)

// ExecutionCompleteToSnapshot transforms the ExecutionComplete event into a SteampipeSnapshot
func ExecutionCompleteToSnapshot(event *dashboardevents.ExecutionComplete) *dashboardtypes.SteampipeSnapshot {
	return &dashboardtypes.SteampipeSnapshot{
		SchemaVersion: fmt.Sprintf("%d", dashboardtypes.SteampipeSnapshotSchemaVersion),
		Panels:        event.Panels,
		Layout:        event.Root.AsTreeNode(),
		Inputs:        event.Inputs,
		Variables:     event.Variables,
		SearchPath:    event.SearchPath,
		StartTime:     event.StartTime,
		EndTime:       event.EndTime,
	}
}
