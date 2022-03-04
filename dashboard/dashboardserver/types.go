package dashboardserver

import (
	"context"
	"sync"

	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/workspace"
	"gopkg.in/olahol/melody.v1"
)

type ListenType string
type ListenPort int

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
}

type InputValuesClearedPayload struct {
	Action        string   `json:"action"`
	ClearedInputs []string `json:"cleared_inputs"`
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
