package dashboardtypes

import (
	"encoding/json"
	steampipecloud "github.com/turbot/steampipe-cloud-sdk-go"
	"time"
)

var SteampipeSnapshotSchemaVersion int64 = 20220929

type SteampipeSnapshot struct {
	SchemaVersion string                   `json:"schema_version"`
	Panels        map[string]SnapshotPanel `json:"panels"`
	Inputs        map[string]interface{}   `json:"inputs"`
	Variables     map[string]string        `json:"variables"`
	SearchPath    []string                 `json:"search_path"`
	StartTime     time.Time                `json:"start_time"`
	EndTime       time.Time                `json:"end_time"`
	Layout        *SnapshotTreeNode        `json:"layout"`
}

// IsExportSourceData implements ExportSourceData
func (*SteampipeSnapshot) IsExportSourceData() {}

func (s *SteampipeSnapshot) AsCloudSnapshot() (*steampipecloud.WorkspaceSnapshotData, error) {
	jsonbytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	res := &steampipecloud.WorkspaceSnapshotData{}
	if err := json.Unmarshal(jsonbytes, res); err != nil {
		return nil, err
	}
	return res, nil
}
