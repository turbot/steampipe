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
	FileNameRoot  string                   `json:"-"`
	Title         string                   `json:"-"`
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

func (s *SteampipeSnapshot) AsStrippedJson(indent bool) ([]byte, error) {
	res, err := s.AsCloudSnapshot()
	if err != nil {
		return nil, err
	}
	if err = StripSnapshot(res); err != nil {
		return nil, err
	}
	if indent {
		return json.MarshalIndent(res, "", "  ")
	}
	return json.Marshal(res)
}

func StripSnapshot(snapshot *steampipecloud.WorkspaceSnapshotData) error {
	propertiesToStrip := []string{
		"sql",
		"source_definition",
		"documentation",
		"search_path",
		"search_path_prefix"}
	for _, p := range snapshot.Panels {
		panel := p.(map[string]any)
		properties, _ := panel["properties"].(map[string]any)
		for _, property := range propertiesToStrip {
			// look both at top level and under properties
			delete(panel, property)
			if properties != nil {
				delete(properties, property)
			}
		}
	}
	// clear top level search path
	snapshot.SearchPath = []string{}
	return nil
}
