package dashboardserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/turbot/go-kit/helpers"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardevents"
	"github.com/turbot/steampipe/pkg/dashboard/dashboardexecute"
	"github.com/turbot/steampipe/pkg/db/db_common"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/pkg/workspace"
	"gopkg.in/olahol/melody.v1"
)

type Server struct {
	context          context.Context
	dbClient         db_common.Client
	mutex            *sync.Mutex
	dashboardClients map[string]*DashboardClientInfo
	webSocket        *melody.Melody
	workspace        *workspace.Workspace
}

func NewServer(ctx context.Context, dbClient db_common.Client, w *workspace.Workspace) (*Server, error) {
	initLogSink()

	OutputWait(ctx, "Starting Dashboard Server")

	webSocket := melody.New()

	var dashboardClients = make(map[string]*DashboardClientInfo)

	var mutex = &sync.Mutex{}

	server := &Server{
		context:          ctx,
		dbClient:         dbClient,
		mutex:            mutex,
		dashboardClients: dashboardClients,
		webSocket:        webSocket,
		workspace:        w,
	}

	w.RegisterDashboardEventHandler(server.HandleDashboardEvent)
	err := w.SetupWatcher(ctx, dbClient, func(c context.Context, e error) {})
	OutputMessage(ctx, "Workspace loaded")

	return server, err
}

// Start starts the API server
// it returns a channel which is signalled when the API server terminates
func (s *Server) Start() chan struct{} {
	s.initAsync(s.context)
	return startAPIAsync(s.context, s.webSocket)
}

// Shutdown stops the API server
func (s *Server) Shutdown() {
	log.Println("[TRACE] Server shutdown")

	if s.webSocket != nil {
		log.Println("[TRACE] closing websocket")
		if err := s.webSocket.Close(); err != nil {
			error_helpers.ShowErrorWithMessage(s.context, err, "Websocket shutdown failed")
		}
		log.Println("[TRACE] closed websocket")
	}

	log.Println("[TRACE] Server shutdown complete")

}

func (s *Server) HandleDashboardEvent(event dashboardevents.DashboardEvent) {
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

	case *dashboardevents.WorkspaceError:
		log.Printf("[TRACE] WorkspaceError event: %s", e.Error)
		payload, payloadError = buildWorkspaceErrorPayload(e)
		if payloadError != nil {
			return
		}
		_ = s.webSocket.Broadcast(payload)
		OutputError(s.context, e.Error)

	case *dashboardevents.ExecutionStarted:
		log.Printf("[TRACE] ExecutionStarted event session %s, dashboard %s", e.Session, e.Root.GetName())
		payload, payloadError = buildExecutionStartedPayload(e)
		if payloadError != nil {
			return
		}
		s.writePayloadToSession(e.Session, payload)
		OutputWait(s.context, fmt.Sprintf("Dashboard execution started: %s", e.Root.GetName()))

	case *dashboardevents.ExecutionError:
		log.Println("[TRACE] execution error event", *e)
		payload, payloadError = buildExecutionErrorPayload(e)
		if payloadError != nil {
			return
		}

		s.writePayloadToSession(e.Session, payload)
		OutputError(s.context, e.Error)

	case *dashboardevents.ExecutionComplete:
		log.Println("[TRACE] execution complete event", *e)
		payload, payloadError = buildExecutionCompletePayload(e)
		if payloadError != nil {
			return
		}
		dashboardName := e.Root.GetName()
		s.writePayloadToSession(e.Session, payload)
		outputReady(s.context, fmt.Sprintf("Execution complete: %s", dashboardName))

	case *dashboardevents.LeafNodeError:
		log.Printf("[TRACE] LeafNodeError event session %s, node %s, error %v", e.Session, e.LeafNode.GetName(), e.Error)

	case *dashboardevents.ControlComplete:
		log.Printf("[TRACE] ControlComplete event session %s, control %s", e.Session, e.Control.GetControlId())
		payload, payloadError = buildControlCompletePayload(e)
		if payloadError != nil {
			return
		}
		s.writePayloadToSession(e.Session, payload)

	case *dashboardevents.ControlError:
		log.Printf("[TRACE] ControlError event session %s, control %s", e.Session, e.Control.GetControlId())
		payload, payloadError = buildControlErrorPayload(e)
		if payloadError != nil {
			return
		}
		s.writePayloadToSession(e.Session, payload)

	case *dashboardevents.LeafNodeComplete:
		log.Printf("[TRACE] LeafNodeComplete event session %s, node %s", e.Session, e.LeafNode.GetName())
		payload, payloadError = buildLeafNodeCompletePayload(e)
		if payloadError != nil {
			return
		}
		s.writePayloadToSession(e.Session, payload)

	case *dashboardevents.DashboardChanged:
		log.Println("[TRACE] DashboardChanged event", *e)
		deletedDashboards := e.DeletedDashboards
		newDashboards := e.NewDashboards

		changedBenchmarks := e.ChangedBenchmarks
		changedCategories := e.ChangedCategories
		changedContainers := e.ChangedContainers
		changedControls := e.ChangedControls
		changedCards := e.ChangedCards
		changedCharts := e.ChangedCharts
		changedDashboards := e.ChangedDashboards
		changedEdges := e.ChangedEdges
		changedFlows := e.ChangedFlows
		changedGraphs := e.ChangedGraphs
		changedHierarchies := e.ChangedHierarchies
		changedImages := e.ChangedImages
		changedInputs := e.ChangedInputs
		changedNodes := e.ChangedNodes
		changedTables := e.ChangedTables
		changedTexts := e.ChangedTexts

		// If nothing has changed, ignore
		if len(deletedDashboards) == 0 &&
			len(newDashboards) == 0 &&
			len(changedBenchmarks) == 0 &&
			len(changedCategories) == 0 &&
			len(changedContainers) == 0 &&
			len(changedControls) == 0 &&
			len(changedCards) == 0 &&
			len(changedCharts) == 0 &&
			len(changedDashboards) == 0 &&
			len(changedEdges) == 0 &&
			len(changedFlows) == 0 &&
			len(changedGraphs) == 0 &&
			len(changedHierarchies) == 0 &&
			len(changedImages) == 0 &&
			len(changedInputs) == 0 &&
			len(changedNodes) == 0 &&
			len(changedTables) == 0 &&
			len(changedTexts) == 0 {
			return
		}

		for k, v := range s.dashboardClients {
			log.Printf("[TRACE] Dashboard client: %v %v\n", k, typeHelpers.SafeString(v.Dashboard))
		}

		// If) any deleted/new/changed dashboards, emit an available dashboards message to clients
		if len(deletedDashboards) != 0 || len(newDashboards) != 0 || len(changedDashboards) != 0 || len(changedBenchmarks) != 0 {
			OutputMessage(s.context, "Available Dashboards updated")
			payload, payloadError = buildAvailableDashboardsPayload(s.workspace.GetResourceMaps())
			if payloadError != nil {
				return
			}
			_ = s.webSocket.Broadcast(payload)
		}

		var dashboardssBeingWatched []string

		dashboardClients := s.getDashboardClients()
		for _, dashboardClientInfo := range dashboardClients {
			dashboardName := typeHelpers.SafeString(dashboardClientInfo.Dashboard)
			if dashboardClientInfo.Dashboard != nil {
				if helpers.StringSliceContains(dashboardssBeingWatched, dashboardName) {
					continue
				}
				dashboardssBeingWatched = append(dashboardssBeingWatched, dashboardName)
			}
		}

		var changedDashboardNames []string
		var newDashboardNames []string

		// Process the changed items and make a note of the dashboard(s) they're in
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedBenchmarks)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedCategories)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedContainers)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedControls)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedCards)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedCharts)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedEdges)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedFlows)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedGraphs)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedHierarchies)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedImages)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedInputs)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedNodes)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedTables)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedTexts)...)

		for _, changedDashboard := range changedDashboards {
			if helpers.StringSliceContains(changedDashboardNames, changedDashboard.Name) {
				continue
			}
			changedDashboardNames = append(changedDashboardNames, changedDashboard.Name)
		}

		for _, changedDashboardName := range changedDashboardNames {
			sessionMap := s.getDashboardClients()
			for sessionId, dashboardClientInfo := range sessionMap {
				if typeHelpers.SafeString(dashboardClientInfo.Dashboard) == changedDashboardName {
					// 					outputMessage(s.context, fmt.Sprintf("Dashboard Changed - executing with inputs: %v", dashboardClientInfo.DashboardInputs))
					_ = dashboardexecute.Executor.ExecuteDashboard(s.context, sessionId, changedDashboardName, dashboardClientInfo.DashboardInputs, s.workspace, s.dbClient)
				}
			}
		}

		// Special case - if we previously had a workspace error, any previously existing dashboards
		// will come in here as new, so we need to check if any of those new dashboards are being watched.
		// If so, execute them
		for _, newDashboard := range newDashboards {
			if helpers.StringSliceContains(newDashboardNames, newDashboard.Name()) {
				continue
			}
			newDashboardNames = append(newDashboardNames, newDashboard.Name())
		}

		sessionMap := s.getDashboardClients()
		for _, newDashboardName := range newDashboardNames {
			for sessionId, dashboardClientInfo := range sessionMap {
				if typeHelpers.SafeString(dashboardClientInfo.Dashboard) == newDashboardName {
					// 					outputMessage(s.context, fmt.Sprintf("New Dashboard - executing with inputs: %v", dashboardClientInfo.DashboardInputs))
					_ = dashboardexecute.Executor.ExecuteDashboard(s.context, sessionId, newDashboardName, dashboardClientInfo.DashboardInputs, s.workspace, s.dbClient)
				}
			}
		}

	case *dashboardevents.DashboardError:
		log.Println("[TRACE] dashboard error event", *e)

	case *dashboardevents.DashboardComplete:
		log.Println("[TRACE] dashboard complete event", *e)

	case *dashboardevents.InputValuesCleared:
		log.Println("[TRACE] input values cleared event", *e)

		payload, payloadError = buildInputValuesClearedPayload(e)
		if payloadError != nil {
			return
		}

		dashboardClients := s.getDashboardClients()
		if sessionInfo, ok := dashboardClients[e.Session]; ok {
			for _, clearedInput := range e.ClearedInputs {
				delete(sessionInfo.DashboardInputs, clearedInput)
			}
			// 			outputMessage(s.context, fmt.Sprintf("Input Values Cleared - dashboard inputs updated: %v", sessionInfo.DashboardInputs))
		}
		s.writePayloadToSession(e.Session, payload)
	}
}

func (s *Server) initAsync(ctx context.Context) {
	go func() {
		// Return list of dashboards on connect
		s.webSocket.HandleConnect(func(session *melody.Session) {
			log.Println("[TRACE] client connected")
			s.addSession(session)
		})

		s.webSocket.HandleDisconnect(func(session *melody.Session) {
			log.Println("[TRACE] client disconnected")
			s.clearSession(ctx, session)
		})

		s.webSocket.HandleMessage(s.handleMessageFunc(ctx))
		OutputMessage(ctx, "Initialization complete")
	}()
}

func (s *Server) handleMessageFunc(ctx context.Context) func(session *melody.Session, msg []byte) {
	return func(session *melody.Session, msg []byte) {

		sessionId := s.getSessionId(session)

		var request ClientRequest
		// if we could not decode message - ignore
		err := json.Unmarshal(msg, &request)
		if err != nil {
			log.Printf("[WARN] failed to marshal message: %s", err.Error())
			return
		}

		if request.Action != "keep_alive" {
			log.Println("[TRACE] message", string(msg))
		}

		switch request.Action {
		case "get_dashboard_metadata":
			payload, err := buildDashboardMetadataPayload(s.workspace.GetResourceMaps(), s.workspace.CloudMetadata)
			if err != nil {
				panic(fmt.Errorf("error building payload for get_metadata: %v", err))
			}
			_ = session.Write(payload)
		case "get_available_dashboards":
			payload, err := buildAvailableDashboardsPayload(s.workspace.GetResourceMaps())
			if err != nil {
				panic(fmt.Errorf("error building payload for get_available_dashboards: %v", err))
			}
			_ = session.Write(payload)
		case "select_dashboard":
			s.setDashboardForSession(sessionId, request.Payload.Dashboard.FullName, request.Payload.InputValues)
			_ = dashboardexecute.Executor.ExecuteDashboard(ctx, sessionId, request.Payload.Dashboard.FullName, request.Payload.InputValues, s.workspace, s.dbClient)
		case "select_snapshot":
			snapshotName := request.Payload.Dashboard.FullName
			s.setDashboardForSession(sessionId, snapshotName, request.Payload.InputValues)
			snap, err := dashboardexecute.Executor.LoadSnapshot(ctx, sessionId, snapshotName, s.workspace)
			// TACTICAL- handle with error message
			error_helpers.FailOnError(err)
			// error handling???
			payload, err := buildDisplaySnapshotPayload(snap)
			// TACTICAL- handle with error message
			error_helpers.FailOnError(err)

			s.writePayloadToSession(sessionId, payload)
			outputReady(s.context, fmt.Sprintf("Show snapshot complete: %s", snapshotName))
		case "input_changed":
			s.setDashboardInputsForSession(sessionId, request.Payload.InputValues)
			_ = dashboardexecute.Executor.OnInputChanged(ctx, sessionId, request.Payload.InputValues, request.Payload.ChangedInput)
		case "clear_dashboard":
			s.setDashboardInputsForSession(sessionId, nil)
			dashboardexecute.Executor.CancelExecutionForSession(ctx, sessionId)
		}
	}
}

func (s *Server) clearSession(ctx context.Context, session *melody.Session) {
	if strings.ToUpper(os.Getenv("DEBUG")) == "TRUE" {
		return
	}

	sessionId := s.getSessionId(session)

	dashboardexecute.Executor.CancelExecutionForSession(ctx, sessionId)

	s.deleteDashboardClient(sessionId)
}

func (s *Server) addSession(session *melody.Session) {
	sessionId := s.getSessionId(session)

	clientSession := &DashboardClientInfo{
		Session: session,
	}

	s.addDashboardClient(sessionId, clientSession)
}

func (s *Server) setDashboardInputsForSession(sessionId string, inputs map[string]interface{}) {
	dashboardClients := s.getDashboardClients()
	if sessionInfo, ok := dashboardClients[sessionId]; ok {
		sessionInfo.DashboardInputs = inputs
		// 		outputMessage(s.context, fmt.Sprintf("Set Dashboard Inputs For Session: %v", sessionInfo.DashboardInputs))
	}
}

func (s *Server) getSessionId(session *melody.Session) string {
	return fmt.Sprintf("%p", session)
}

// functions providing locked access to member properties

func (s *Server) setDashboardForSession(sessionId string, dashboardName string, inputs map[string]interface{}) *DashboardClientInfo {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	dashboardClientInfo := s.dashboardClients[sessionId]
	dashboardClientInfo.Dashboard = &dashboardName
	dashboardClientInfo.DashboardInputs = inputs
	//outputMessage(s.context, fmt.Sprintf("Set Dashboard For Session - initial inputs: %v", dashboardClientInfo.DashboardInputs))

	return dashboardClientInfo
}

func (s *Server) writePayloadToSession(sessionId string, payload []byte) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if sessionInfo, ok := s.dashboardClients[sessionId]; ok {
		_ = sessionInfo.Session.Write(payload)
	}
}

func (s *Server) getDashboardClients() map[string]*DashboardClientInfo {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.dashboardClients
}

func (s *Server) addDashboardClient(sessionId string, clientSession *DashboardClientInfo) {
	s.mutex.Lock()
	s.dashboardClients[sessionId] = clientSession
	s.mutex.Unlock()
}

func (s *Server) deleteDashboardClient(sessionId string) {
	s.mutex.Lock()
	delete(s.dashboardClients, sessionId)
	s.mutex.Unlock()
}

func getDashboardsInterestedInResourceChanges(dashboardsBeingWatched []string, existingChangedDashboardNames []string, changedItems []*modconfig.DashboardTreeItemDiffs) []string {
	var changedDashboardNames []string

	for _, changedItem := range changedItems {
		paths := changedItem.Item.GetPaths()
		for _, nodePath := range paths {
			for _, nodeName := range nodePath {
				resourceParts, _ := modconfig.ParseResourceName(nodeName)
				// We only care about changes from these resource types
				if !helpers.StringSliceContains([]string{modconfig.BlockTypeDashboard, modconfig.BlockTypeBenchmark}, resourceParts.ItemType) {
					continue
				}

				if helpers.StringSliceContains(existingChangedDashboardNames, nodeName) || helpers.StringSliceContains(changedDashboardNames, nodeName) || !helpers.StringSliceContains(dashboardsBeingWatched, nodeName) {
					continue
				}

				changedDashboardNames = append(changedDashboardNames, nodeName)
			}
		}
	}

	return changedDashboardNames
}
