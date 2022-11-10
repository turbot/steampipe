package dashboardserver

import (
	"fmt"
	"github.com/turbot/steampipe/pkg/control/controlstatus"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/steampipeconfig"
	"gopkg.in/olahol/melody.v1"
)

type ListenType string

const (
	ListenTypeLocal   ListenType = "local"
	ListenTypeNetwork ListenType = "network"
)

// IsValid is a validator for ListenType known values
func (lt ListenType) IsValid() error {
	switch lt {
	case ListenTypeNetwork, ListenTypeLocal:
		return nil
	}
	return fmt.Errorf("invalid listen type. Must be one of '%v' or '%v'", ListenTypeNetwork, ListenTypeLocal)
}

type ListenPort int

// IsValid is a validator for ListenType known values
func (lp ListenPort) IsValid() error {
	if lp < 1 || lp > 65535 {
		return fmt.Errorf("invalid port - must be within range (1:65535)")
	}
	return nil
}

type ErrorPayload struct {
	Action string `json:"action"`
	Error  string `json:"error"`
}

var ExecutionStartedSchemaVersion int64 = 20220614

type ExecutionStartedPayload struct {
	SchemaVersion string                                  `json:"schema_version"`
	Action        string                                  `json:"action"`
	ExecutionId   string                                  `json:"execution_id"`
	Panels        map[string]dashboardtypes.SnapshotPanel `json:"panels"`
	Layout        *dashboardtypes.SnapshotTreeNode        `json:"layout"`
	Inputs        map[string]interface{}                  `json:"inputs,omitempty"`
	Variables     map[string]string                       `json:"variables,omitempty"`
}

type LeafNodeCompletePayload struct {
	Action        string                          `json:"action"`
	DashboardNode dashboardtypes.DashboardNodeRun `json:"dashboard_node"`
	ExecutionId   string                          `json:"execution_id"`
}

type ControlEventPayload struct {
	Action      string                                 `json:"action"`
	Control     controlstatus.ControlRunStatusProvider `json:"control"`
	Name        string                                 `json:"name"`
	Progress    *controlstatus.ControlProgress         `json:"progress"`
	ExecutionId string                                 `json:"execution_id"`
}

type ExecutionErrorPayload struct {
	Action string `json:"action"`
	Error  string `json:"error"`
}

var ExecutionCompletePayloadSchemaVersion int64 = 20220929

type ExecutionCompletePayload struct {
	Action        string                            `json:"action"`
	SchemaVersion string                            `json:"schema_version"`
	Snapshot      *dashboardtypes.SteampipeSnapshot `json:"snapshot"`
	ExecutionId   string                            `json:"execution_id"`
}

type DisplaySnapshotPayload struct {
	Action        string `json:"action"`
	SchemaVersion string `json:"schema_version"`
	// snapshot is a map here as we cannot deserialise SteampipeSnapshot into a struct
	// (without custom derserialisation code) as the Panels property is an interface
	Snapshot    map[string]any `json:"snapshot"`
	ExecutionId string         `json:"execution_id"`
}

type InputValuesClearedPayload struct {
	Action        string   `json:"action"`
	ClearedInputs []string `json:"cleared_inputs"`
	ExecutionId   string   `json:"execution_id"`
}

type DashboardClientInfo struct {
	Session         *melody.Session
	Dashboard       *string
	DashboardInputs map[string]interface{}
}

type ClientRequestDashboardPayload struct {
	FullName string `json:"full_name"`
}

type ClientRequestPayload struct {
	Dashboard    ClientRequestDashboardPayload `json:"dashboard"`
	InputValues  map[string]interface{}        `json:"input_values"`
	ChangedInput string                        `json:"changed_input"`
}

type ClientRequest struct {
	Action  string               `json:"action"`
	Payload ClientRequestPayload `json:"payload"`
}

type ModAvailableDashboard struct {
	Title       string            `json:"title,omitempty"`
	FullName    string            `json:"full_name"`
	ShortName   string            `json:"short_name"`
	Tags        map[string]string `json:"tags"`
	ModFullName string            `json:"mod_full_name"`
}

type ModAvailableBenchmark struct {
	Title       string                  `json:"title,omitempty"`
	FullName    string                  `json:"full_name"`
	ShortName   string                  `json:"short_name"`
	Tags        map[string]string       `json:"tags"`
	IsTopLevel  bool                    `json:"is_top_level"`
	Children    []ModAvailableBenchmark `json:"children,omitempty"`
	Trunks      [][]string              `json:"trunks"`
	ModFullName string                  `json:"mod_full_name"`
}

type AvailableDashboardsPayload struct {
	Action     string                           `json:"action"`
	Dashboards map[string]ModAvailableDashboard `json:"dashboards"`
	Benchmarks map[string]ModAvailableBenchmark `json:"benchmarks"`
	Snapshots  map[string]string                `json:"snapshots"`
}

type ModDashboardMetadata struct {
	Title     string `json:"title,omitempty"`
	FullName  string `json:"full_name"`
	ShortName string `json:"short_name"`
}

type DashboardCLIMetadata struct {
	Version string `json:"version,omitempty"`
}

type DashboardMetadata struct {
	Mod           *ModDashboardMetadata           `json:"mod,omitempty"`
	InstalledMods map[string]ModDashboardMetadata `json:"installed_mods,omitempty"`
	CLI           DashboardCLIMetadata            `json:"cli"`
	Cloud         *steampipeconfig.CloudMetadata  `json:"cloud,omitempty"`
	Telemetry     string                          `json:"telemetry"`
}

type DashboardMetadataPayload struct {
	Action   string            `json:"action"`
	Metadata DashboardMetadata `json:"metadata"`
}
