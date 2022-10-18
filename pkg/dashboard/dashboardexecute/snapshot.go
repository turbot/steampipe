package dashboardexecute

import (
	"context"
	"fmt"
	"log"

	"github.com/turbot/steampipe/pkg/contexthelpers"
	"github.com/turbot/steampipe/pkg/control/controlstatus"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/initialisation"
	"github.com/turbot/steampipe/pkg/statushooks"
)

func GenerateSnapshot(ctx context.Context, target string, initData *initialisation.InitData, inputs map[string]interface{}) (snapshot *dashboardtypes.SteampipeSnapshot, err error) {
	defer statushooks.Done(ctx)
	// create context for the dashboard execution
	snapshotCtx := createSnapshotContext(ctx, target)

	w := initData.Workspace

	// no session for manual execution
	sessionId := ""

	errorChannel := make(chan error)
	resultChannel := make(chan *dashboardtypes.SteampipeSnapshot)
	dashboardEventHandler := func(event dashboardevents.DashboardEvent) {
		handleDashboardEvent(event, resultChannel, errorChannel)
	}
	w.RegisterDashboardEventHandler(dashboardEventHandler)
	Executor.ExecuteDashboard(snapshotCtx, sessionId, target, inputs, w, initData.Client)

	select {
	case err = <-errorChannel:
	case snapshot = <-resultChannel:
	}
	// clear event handlers again in case another snapshot will be generated in this run
	w.UnregisterDashboardEventHandlers()

	return snapshot, err
}

// create the context for the check run - add a control status renderer
func createSnapshotContext(ctx context.Context, target string) context.Context {
	// create context for the dashboard execution
	snapshotCtx, cancel := context.WithCancel(ctx)
	contexthelpers.StartCancelHandler(cancel)

	snapshotProgressReporter := statushooks.NewSnapshotProgressReporter(target)
	snapshotCtx = statushooks.AddSnapshotProgressToContext(snapshotCtx, snapshotProgressReporter)

	// create a context with a SnapshotControlHooks to report execution progress of any controls in this snapshot
	snapshotCtx = controlstatus.AddControlHooksToContext(snapshotCtx, controlstatus.NewSnapshotControlHooks())
	return snapshotCtx
}

func handleDashboardEvent(event dashboardevents.DashboardEvent, resultChannel chan *dashboardtypes.SteampipeSnapshot, errorChannel chan error) {
	switch e := event.(type) {
	case *dashboardevents.ExecutionError:
		errorChannel <- e.Error
	case *dashboardevents.ExecutionComplete:
		log.Println("[TRACE] execution complete event", *e)
		snap := ExecutionCompleteToSnapshot(e)
		resultChannel <- snap
	}
}

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
