package snapshot

import (
	"context"
	"fmt"
	"github.com/turbot/steampipe/pkg/contexthelpers"
	"github.com/turbot/steampipe/pkg/dashboard"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardexecute"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardserver"
	"github.com/turbot/steampipe/pkg/interactive"
	"github.com/turbot/steampipe/pkg/utils"
	"log"
	"reflect"
)

func GenerateSnapshot(target string) (snapshot string, err error) {
	// create context for the dashboard execution
	ctx, cancel := context.WithCancel(context.Background())
	contexthelpers.StartCancelHandler(cancel)

	w, err := interactive.LoadWorkspacePromptingForVariables(ctx)
	utils.FailOnErrorWithMessage(err, "failed to load workspace")

	initData := dashboard.NewInitData(ctx, w)
	// shutdown the service on exit
	defer initData.Cleanup(ctx)
	if err := initData.Result.Error; err != nil {
		return "", initData.Result.Error
	}
	// cancelled?
	if ctx != nil && ctx.Err() != nil {
		return "", ctx.Err()
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
	dashboardexecute.Executor.ExecuteDashboard(ctx, sessionId, target, inputs, w, initData.Client)

	select {
	case err = <-errorChannel:
	case snapshot = <-resultChannel:
		// publish if needed
	}

	return snapshot, err

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
		payload, payloadError = dashboardserver.BuildExecutionCompletePayload(e)
		if payloadError != nil {
			errorChannel <- payloadError
		}
		resultChannel <- string(payload)
	}
}