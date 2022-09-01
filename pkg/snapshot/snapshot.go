package snapshot

import (
	"context"
	"github.com/mattn/go-isatty"
	"github.com/turbot/steampipe/pkg/contexthelpers"
	"github.com/turbot/steampipe/pkg/control/controlstatus"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardexecute"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardserver"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardtypes"
	"github.com/turbot/steampipe/pkg/initialisation"
	"github.com/turbot/steampipe/pkg/interactive"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/utils"
	"log"
	"os"
)

func GenerateSnapshot(ctx context.Context, target string) (snapshot *dashboardtypes.SteampipeSnapshot, err error) {
	// create context for the dashboard execution
	snapshotCtx, cancel := createSnapshotContext(ctx, target)

	contexthelpers.StartCancelHandler(cancel)

	w, err := interactive.LoadWorkspacePromptingForVariables(snapshotCtx)
	utils.FailOnErrorWithMessage(err, "failed to load workspace")

	// todo do we require a mod file?

	initData := initialisation.NewInitData(snapshotCtx, w)
	// shutdown the service on exit
	defer initData.Cleanup(snapshotCtx)
	if err := initData.Result.Error; err != nil {
		return nil, initData.Result.Error
	}

	// if there is a usage warning we display it
	initData.Result.DisplayMessages()

	sessionId := "generateSnapshot"

	// todo KAI get inputs from command line
	inputs := make(map[string]interface{})

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

	var controlHooks controlstatus.ControlHooks = controlstatus.NullHooks
	// TODO KAI only do tty check for actual status spinner
	// if the client is a TTY, inject a status spinner
	if isatty.IsTerminal(os.Stdout.Fd()) {
		controlHooks = controlstatus.NewSnapshotControlHooks()
	}

	// create a context with a SnapshotControlHooks to report execution progress of any controls in this snapshot
	snapshotCtx = controlstatus.AddControlHooksToContext(snapshotCtx, controlHooks)
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
