package snapshot

import (
	"context"
	"fmt"
	"github.com/mattn/go-isatty"
	"github.com/turbot/steampipe/pkg/contexthelpers"
	"github.com/turbot/steampipe/pkg/control/controlstatus"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardexecute"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardserver"
	"github.com/turbot/steampipe/pkg/initialisation"
	"github.com/turbot/steampipe/pkg/interactive"
	"github.com/turbot/steampipe/pkg/statushooks"
	"github.com/turbot/steampipe/pkg/utils"
	"log"
	"os"
	"reflect"
)

func GenerateSnapshot(ctx context.Context, target string) (snapshot string, err error) {
	// TODO WHAT AM I??
	snapshotAddress := "http://snapshot/address"
	// create context for the dashboard execution
	snapshotCtx, cancel := createSnapshotContext(ctx, target, snapshotAddress)
	
	contexthelpers.StartCancelHandler(cancel)

	w, err := interactive.LoadWorkspacePromptingForVariables(snapshotCtx)
	utils.FailOnErrorWithMessage(err, "failed to load workspace")

	// todo do we require a mod file?

	initData := initialisation.NewInitData(snapshotCtx, w)
	// shutdown the service on exit
	defer initData.Cleanup(snapshotCtx)
	if err := initData.Result.Error; err != nil {
		return "", initData.Result.Error
	}

	// if there is a usage warning we display it
	initData.Result.DisplayMessages()

	sessionId := "generateSnapshot"

	// todo get inputs from command line
	inputs := make(map[string]interface{})

	errorChannel := make(chan error)
	resultChannel := make(chan string)
	dashboardEventHandler := func(event dashboardevents.DashboardEvent) {
		handleDashboardEvent(event, resultChannel, errorChannel)
	}
	w.RegisterDashboardEventHandler(dashboardEventHandler)
	dashboardexecute.Executor.ExecuteDashboard(snapshotCtx, sessionId, target, inputs, w, initData.Client)

	select {
	case err = <-errorChannel:
	case snapshot = <-resultChannel:
		// publish if needed
	}

	return snapshot, err

}

// create the context for the check run - add a control status renderer
func createSnapshotContext(ctx context.Context, target string, snapshotAddress string) (context.Context, context.CancelFunc) {
	// create context for the dashboard execution
	snapshotCtx, cancel := context.WithCancel(ctx)
	contexthelpers.StartCancelHandler(cancel)

	snapshotProgressReporter := NewSnapshotProgressReporter(target, snapshotAddress)
	snapshotCtx = statushooks.AddSnapshotProgressToContext(snapshotCtx, snapshotProgressReporter)

	var controlHooks controlstatus.ControlHooks = controlstatus.NullHooks
	// TODO only do tty check for actual status spinner
	// if the client is a TTY, inject a status spinner
	if isatty.IsTerminal(os.Stdout.Fd()) {
		controlHooks = controlstatus.NewSnapshotControlHooks()
	}

	// create a context with a SnapshotControlHooks to report execution progress of any controls in this snapshot
	snapshotCtx = controlstatus.AddControlHooksToContext(snapshotCtx, controlHooks)
	return snapshotCtx, cancel
}

func handleDashboardEvent(event dashboardevents.DashboardEvent, resultChannel chan string, errorChannel chan error) {
	var payloadError error
	var payload []byte
	defer func() {
		if payloadError != nil {
			// we don't expect the build functions to ever error during marshalling
			// this is because the data getting marshalled are not expected to have go specific
			// properties/data in them
			panic(fmt.Errorf("error building payload for '%s': %v", reflect.TypeOf(event).String(), payloadError))
		}
	}()

	switch e := event.(type) {

	case *dashboardevents.ExecutionError:
		errorChannel <- e.Error
	case *dashboardevents.ExecutionComplete:
		log.Println("[TRACE] execution complete event", *e)
		payload, payloadError = dashboardserver.BuildExecutionCompletePayload(e, true)
		if payloadError != nil {
			errorChannel <- payloadError
		}
		resultChannel <- string(payload)
	}
}
