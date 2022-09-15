package snapshot

import (
	"context"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/contexthelpers"
	"github.com/turbot/steampipe/pkg/control/controlstatus"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardexecute"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardserver"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/initialisation"
	"github.com/turbot/steampipe/pkg/interactive"
	"github.com/turbot/steampipe/pkg/statushooks"
	"log"
)

func GenerateSnapshot(ctx context.Context, target string, inputs map[string]interface{}) (snapshot *dashboardtypes.SteampipeSnapshot, err error) {
	// create context for the dashboard execution
	snapshotCtx, cancel := createSnapshotContext(ctx, target)

	contexthelpers.StartCancelHandler(cancel)

	w, err := interactive.LoadWorkspacePromptingForVariables(snapshotCtx)
	error_helpers.FailOnErrorWithMessage(err, "failed to load workspace")

	// todo do we require a mod file?

	initData := initialisation.NewInitData(snapshotCtx, w, constants.InvokerDashboard)
	// shutdown the service on exit
	defer initData.Cleanup(snapshotCtx)
	if err := initData.Result.Error; err != nil {
		return nil, initData.Result.Error
	}

	// if there is a usage warning we display it
	initData.Result.DisplayMessages(nil)

	sessionId := "generateSnapshot"

	errorChannel := make(chan error)
	resultChannel := make(chan *dashboardtypes.SteampipeSnapshot)
	dashboardEventHandler := func(event dashboardevents.DashboardEvent) {
		handleDashboardEvent(event, resultChannel, errorChannel)
	}
	w.RegisterDashboardEventHandler(dashboardEventHandler)
	dashboardexecute.Executor.ExecuteDashboard(snapshotCtx, sessionId, target, inputs, w, initData.Client)

	select {
	case err = <-errorChannel:
	case snapshot = <-resultChannel:
	}

	return snapshot, err

}

// create the context for the check run - add a control status renderer
func createSnapshotContext(ctx context.Context, target string) (context.Context, context.CancelFunc) {
	// create context for the dashboard execution
	snapshotCtx, cancel := context.WithCancel(ctx)
	contexthelpers.StartCancelHandler(cancel)

	snapshotProgressReporter := NewSnapshotProgressReporter(target)
	snapshotCtx = statushooks.AddSnapshotProgressToContext(snapshotCtx, snapshotProgressReporter)

	// create a context with a SnapshotControlHooks to report execution progress of any controls in this snapshot
	snapshotCtx = controlstatus.AddControlHooksToContext(snapshotCtx, controlstatus.NewSnapshotControlHooks())
	return snapshotCtx, cancel
}

func handleDashboardEvent(event dashboardevents.DashboardEvent, resultChannel chan *dashboardtypes.SteampipeSnapshot, errorChannel chan error) {

	switch e := event.(type) {

	case *dashboardevents.ExecutionError:
		errorChannel <- e.Error
	case *dashboardevents.ExecutionComplete:
		log.Println("[TRACE] execution complete event", *e)
		snapshot := dashboardserver.ExecutionCompleteToSnapshot(e)

		resultChannel <- snapshot
	}
}
