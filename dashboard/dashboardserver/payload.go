package dashboardserver

import (
	"encoding/json"

	"github.com/spf13/viper"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/dashboard/dashboardevents"
	"github.com/turbot/steampipe/steampipeconfig"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

func buildDashboardMetadataPayload(workspaceResources *modconfig.ModResources, cloudMetadata *steampipeconfig.CloudMetadata) ([]byte, error) {
	installedMods := make(map[string]ModDashboardMetadata)
	for _, mod := range workspaceResources.Mods {
		// Ignore current mod
		if mod.FullName == workspaceResources.Mod.FullName {
			continue
		}
		installedMods[mod.FullName] = ModDashboardMetadata{
			Title:     typeHelpers.SafeString(mod.Title),
			FullName:  mod.FullName,
			ShortName: mod.ShortName,
		}
	}

	payload := DashboardMetadataPayload{
		Action: "dashboard_metadata",
		Metadata: DashboardMetadata{
			Mod: ModDashboardMetadata{
				Title:     typeHelpers.SafeString(workspaceResources.Mod.Title),
				FullName:  workspaceResources.Mod.FullName,
				ShortName: workspaceResources.Mod.ShortName,
			},
			InstalledMods: installedMods,
			Telemetry:     viper.GetString(constants.ArgTelemetry),
		},
	}

	// if telemetry is enabled, send cloud metadata
	if payload.Metadata.Telemetry != constants.TelemetryNone {
		payload.Metadata.Cloud = cloudMetadata
	}

	return json.Marshal(payload)
}

func buildAvailableDashboardsPayload(workspaceResources *modconfig.ModResources) ([]byte, error) {
	// build a map of the dashboards provided by each mod
	dashboardsByMod := make(map[string]map[string]ModAvailableDashboard)

	// iterate over the dashboards for the top level mod - this will include the dashboards from dependency mods
	for _, dashboard := range workspaceResources.Mod.ResourceMaps.Dashboards {
		mod := dashboard.Mod
		// create a child map for this mod if needed
		if _, ok := dashboardsByMod[mod.FullName]; !ok {
			dashboardsByMod[mod.FullName] = make(map[string]ModAvailableDashboard)
		}
		// add this dashboard
		dashboardsByMod[mod.FullName][dashboard.FullName] = ModAvailableDashboard{
			Title:     typeHelpers.SafeString(dashboard.Title),
			FullName:  dashboard.FullName,
			ShortName: dashboard.ShortName,
			Tags:      dashboard.Tags,
		}
	}
	for _, benchmark := range workspaceResources.Mod.ResourceMaps.Benchmarks {
		if benchmark.IsAnonymous() {
			continue
		}
		mod := benchmark.Mod
		// create a child map for this mod if needed
		if _, ok := dashboardsByMod[mod.FullName]; !ok {
			dashboardsByMod[mod.FullName] = make(map[string]ModAvailableDashboard)
		}
		// add this dashboard
		dashboardsByMod[mod.FullName][benchmark.FullName] = ModAvailableDashboard{
			Title:     typeHelpers.SafeString(benchmark.Title),
			FullName:  benchmark.FullName,
			ShortName: benchmark.ShortName,
			Tags:      benchmark.Tags,
		}
	}
	payload := AvailableDashboardsPayload{
		Action:          "available_dashboards",
		DashboardsByMod: dashboardsByMod,
	}
	return json.Marshal(payload)
}

func buildWorkspaceErrorPayload(e *dashboardevents.WorkspaceError) ([]byte, error) {
	payload := ErrorPayload{
		Action: "workspace_error",
		Error:  e.Error.Error(),
	}
	return json.Marshal(payload)
}

func buildControlCompletePayload(event *dashboardevents.ControlComplete) ([]byte, error) {
	payload := ControlEventPayload{
		Action:               "control_complete",
		ControlName:          event.ControlName,
		ControlStatusSummary: event.ControlStatusSummary,
		ControlRunStatus:     event.ControlRunStatus,
		Progress:             event.Progress,
		ExecutionId:          event.ExecutionId,
	}
	return json.Marshal(payload)
}
func buildControlErrorPayload(event *dashboardevents.ControlError) ([]byte, error) {
	payload := ControlEventPayload{
		Action:               "control_error",
		ControlName:          event.ControlName,
		ControlStatusSummary: event.ControlStatusSummary,
		ControlRunStatus:     event.ControlRunStatus,
		Progress:             event.Progress,
		ExecutionId:          event.ExecutionId,
	}
	return json.Marshal(payload)
}

func buildLeafNodeCompletePayload(event *dashboardevents.LeafNodeComplete) ([]byte, error) {
	payload := ExecutionPayload{
		Action:        "leaf_node_complete",
		DashboardNode: event.LeafNode,
		ExecutionId:   event.ExecutionId,
	}
	return json.Marshal(payload)
}

func buildExecutionStartedPayload(event *dashboardevents.ExecutionStarted) ([]byte, error) {
	payload := ExecutionPayload{
		Action:        "execution_started",
		DashboardNode: event.Root,
		ExecutionId:   event.ExecutionId,
	}
	return json.Marshal(payload)
}

func buildExecutionCompletePayload(event *dashboardevents.ExecutionComplete) ([]byte, error) {
	payload := ExecutionPayload{
		Action:        "execution_complete",
		DashboardNode: event.Root,
		ExecutionId:   event.ExecutionId,
	}
	return json.Marshal(payload)
}

func buildInputValuesClearedPayload(event *dashboardevents.InputValuesCleared) ([]byte, error) {
	payload := InputValuesClearedPayload{
		Action:        "input_values_cleared",
		ClearedInputs: event.ClearedInputs,
		ExecutionId:   event.ExecutionId,
	}
	return json.Marshal(payload)
}
