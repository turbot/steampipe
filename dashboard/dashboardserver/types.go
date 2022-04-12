package dashboardserver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/workspace"
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

type Server struct {
	context          context.Context
	dbClient         db_common.Client
	mutex            *sync.Mutex
	dashboardClients map[string]*DashboardClientInfo
	webSocket        *melody.Melody
	workspace        *workspace.Workspace
}

type ErrorPayload struct {
	Action string `json:"action"`
	Error  string `json:"error"`
}

type ExecutionPayload struct {
	Action        string                               `json:"action"`
	DashboardNode dashboardinterfaces.DashboardNodeRun `json:"dashboard_node"`
	ExecutionId   string                               `json:"execution_id"`
}

var ExecutionCompleteSchemaVersion int64 = 20220411

type ExecutionCompletePayload struct {
	SchemaVersion int64                                `json:"schema_version"`
	Action        string                               `json:"action"`
	DashboardNode dashboardinterfaces.DashboardNodeRun `json:"dashboard_node"`
	ExecutionId   string                               `json:"execution_id"`
	Inputs        map[string]interface{}               `json:"inputs"`
	Variables     map[string]string                    `json:"variables"`
	SearchPath    []string                             `json:"search_path"`
	StartTime     time.Time                            `json:"start_time"`
	EndTime       time.Time                            `json:"end_time"`
	Actor         string                               `json:"actor"`
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
	Title     string            `json:"title,omitempty"`
	FullName  string            `json:"full_name"`
	ShortName string            `json:"short_name"`
	Tags      map[string]string `json:"tags"`
}

type AvailableDashboardsPayload struct {
	Action          string                                      `json:"action"`
	DashboardsByMod map[string]map[string]ModAvailableDashboard `json:"dashboards_by_mod"`
}

type ModDashboardMetadata struct {
	Title     string `json:"title,omitempty"`
	FullName  string `json:"full_name"`
	ShortName string `json:"short_name"`
}

type DashboardMetadata struct {
	Mod           ModDashboardMetadata            `json:"mod"`
	InstalledMods map[string]ModDashboardMetadata `json:"installed_mods,omitempty"`
	Cloud         *steampipeconfig.CloudMetadata  `json:"cloud,omitempty"`
	Telemetry     string                          `json:"telemetry"`
}

type DashboardMetadataPayload struct {
	Action   string            `json:"action"`
	Metadata DashboardMetadata `json:"metadata"`
}
