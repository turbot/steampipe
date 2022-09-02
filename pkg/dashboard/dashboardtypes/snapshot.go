package dashboardtypes

import (
	"time"
)

var SteampipeSnapshotSchemaVersion int64 = 20220614

type SteampipeSnapshot struct {
	SchemaVersion string                   `json:"schema_version"`
	Action        string                   `json:"action"`
	DashboardNode DashboardNodeRun         `json:"dashboard_node"`
	Panels        map[string]SnapshotPanel `json:"panels"`
	ExecutionId   string                   `json:"execution_id"`
	Inputs        map[string]interface{}   `json:"inputs"`
	Variables     map[string]string        `json:"variables"`
	SearchPath    []string                 `json:"search_path"`
	StartTime     time.Time                `json:"start_time"`
	EndTime       time.Time                `json:"end_time"`
	Layout        *SnapshotTreeNode        `json:"layout"`
}
