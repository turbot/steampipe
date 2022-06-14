package dashboardserver

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v3/instrument"
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
	if payload.Metadata.Telemetry != instrument.TelemetryNone {
		payload.Metadata.Cloud = cloudMetadata
	}

	return json.Marshal(payload)
}

func addBenchmarkChildren(benchmark *modconfig.Benchmark, recordTrunk bool, trunk []string, trunks map[string][][]string) []ModAvailableBenchmark {
	var children []ModAvailableBenchmark
	for _, child := range benchmark.GetChildren() {
		switch t := child.(type) {
		case *modconfig.Benchmark:
			childTrunk := make([]string, len(trunk)+1)
			copy(childTrunk, trunk)
			childTrunk[len(childTrunk)-1] = t.FullName
			if recordTrunk {
				trunks[t.FullName] = append(trunks[t.FullName], childTrunk)
			}
			availableBenchmark := ModAvailableBenchmark{
				Title:     t.GetTitle(),
				FullName:  t.FullName,
				ShortName: t.ShortName,
				Tags:      t.Tags,
				Children:  addBenchmarkChildren(t, recordTrunk, childTrunk, trunks),
			}
			children = append(children, availableBenchmark)
		}
	}
	return children
}

func buildAvailableDashboardsPayload(workspaceResources *modconfig.ModResources) ([]byte, error) {
	// build a map of the dashboards provided by each mod
	dashboards := make(map[string]ModAvailableDashboard)

	// iterate over the dashboards for the top level mod - this will include the dashboards from dependency mods
	for _, dashboard := range workspaceResources.Mod.ResourceMaps.Dashboards {
		mod := dashboard.Mod
		// add this dashboard
		dashboards[dashboard.FullName] = ModAvailableDashboard{
			Title:       typeHelpers.SafeString(dashboard.Title),
			FullName:    dashboard.FullName,
			ShortName:   dashboard.ShortName,
			Tags:        dashboard.Tags,
			ModFullName: mod.FullName,
		}
	}

	benchmarks := make(map[string]ModAvailableBenchmark)
	benchmarkTrunks := make(map[string][][]string)

	for _, benchmark := range workspaceResources.Mod.ResourceMaps.Benchmarks {
		if benchmark.IsAnonymous() {
			continue
		}

		// Find any benchmarks who have a parent that is a mod - we consider these top-level
		isTopLevel := false
		for _, parent := range benchmark.Parents {
			switch parent.(type) {
			case *modconfig.Mod:
				isTopLevel = true
			}
		}

		mod := benchmark.Mod
		trunk := []string{benchmark.FullName}

		if isTopLevel {
			benchmarkTrunks[benchmark.FullName] = [][]string{trunk}
		}

		availableBenchmark := ModAvailableBenchmark{
			Title:       benchmark.GetTitle(),
			FullName:    benchmark.FullName,
			ShortName:   benchmark.ShortName,
			Tags:        benchmark.Tags,
			IsTopLevel:  isTopLevel,
			Children:    addBenchmarkChildren(benchmark, isTopLevel, trunk, benchmarkTrunks),
			ModFullName: mod.FullName,
		}

		benchmarks[benchmark.FullName] = availableBenchmark
	}
	for benchmarkName, trunks := range benchmarkTrunks {
		if foundBenchmark, ok := benchmarks[benchmarkName]; ok {
			foundBenchmark.Trunks = trunks
			benchmarks[benchmarkName] = foundBenchmark
		}
	}

	payload := AvailableDashboardsPayload{
		Action:     "available_dashboards",
		Dashboards: dashboards,
		Benchmarks: benchmarks,
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
		Action:      "control_complete",
		Control:     event.Control,
		Name:        event.Name,
		Progress:    event.Progress,
		ExecutionId: event.ExecutionId,
	}
	return json.Marshal(payload)
}

func buildControlErrorPayload(event *dashboardevents.ControlError) ([]byte, error) {
	payload := ControlEventPayload{
		Action:      "control_error",
		Control:     event.Control,
		Name:        event.Name,
		Progress:    event.Progress,
		ExecutionId: event.ExecutionId,
	}
	return json.Marshal(payload)
}

func buildLeafNodeCompletePayload(event *dashboardevents.LeafNodeComplete) ([]byte, error) {
	payload := LeafNodeCompletePayload{
		Action:        "leaf_node_complete",
		DashboardNode: event.LeafNode,
		ExecutionId:   event.ExecutionId,
	}
	return json.Marshal(payload)
}

func buildExecutionStartedPayload(event *dashboardevents.ExecutionStarted) ([]byte, error) {
	payload := ExecutionStartedPayload{
		SchemaVersion: fmt.Sprintf("%d", ExecutionStartedSchemaVersion),
		Action:        "execution_started",
		DashboardNode: event.Root,
		ExecutionId:   event.ExecutionId,
		Panels:        event.Panels,
		Layout:        event.Root.AsTreeNode(),
	}
	return json.Marshal(payload)
}

func buildExecutionErrorPayload(event *dashboardevents.ExecutionError) ([]byte, error) {
	payload := ExecutionErrorPayload{
		Action: "execution_error",
		Error:  event.Error.Error(),
	}
	return json.Marshal(payload)
}

func buildExecutionCompletePayload(event *dashboardevents.ExecutionComplete) ([]byte, error) {
	payload := ExecutionCompletePayload{
		SchemaVersion: fmt.Sprintf("%d", ExecutionCompleteSchemaVersion),
		Action:        "execution_complete",
		DashboardNode: event.Root,
		ExecutionId:   event.ExecutionId,
		Panels:        event.Panels,
		Layout:        event.Root.AsTreeNode(),
		Inputs:        event.Inputs,
		Variables:     event.Variables,
		SearchPath:    event.SearchPath,
		StartTime:     event.StartTime,
		EndTime:       event.EndTime,
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
