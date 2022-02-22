package dashboardserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/turbot/go-kit/helpers"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/dashboard/dashboardevents"
	"github.com/turbot/steampipe/dashboard/dashboardexecute"
	"github.com/turbot/steampipe/dashboard/dashboardinterfaces"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
	"gopkg.in/olahol/melody.v1"
)

type ListenType string
type ListenPort int

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
}

type DashboardClientInfo struct {
	Session   *melody.Session
	Dashboard *string
}

func NewServer(ctx context.Context, dbClient db_common.Client, w *workspace.Workspace) (*Server, error) {
	outputWait(ctx, "Starting Dashboard Server")

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

	w.RegisterDashboardEventHandler(server.HandleWorkspaceUpdate)
	err := w.SetupWatcher(ctx, dbClient, func(c context.Context, e error) {})
	outputMessage(ctx, "Workspace loaded")

	return server, err
}

func buildDashboardMetadataPayload(workspaceResources *modconfig.WorkspaceResourceMaps) ([]byte, error) {
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
		},
	}
	return json.Marshal(payload)
}

func buildAvailableDashboardsPayload(workspaceResources *modconfig.WorkspaceResourceMaps) ([]byte, error) {
	dashboardsByMod := make(map[string]map[string]ModAvailableDashboard)
	for _, mod := range workspaceResources.Mods {
		_, ok := dashboardsByMod[mod.FullName]
		if !ok {
			dashboardsByMod[mod.FullName] = make(map[string]ModAvailableDashboard)
		}
		for _, dashboard := range mod.Dashboards {
			if dashboard.IsTopLevel {
				dashboardsByMod[mod.FullName][dashboard.FullName] = ModAvailableDashboard{
					Title:     typeHelpers.SafeString(dashboard.Title),
					FullName:  dashboard.FullName,
					ShortName: dashboard.ShortName,
				}
			}
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

func buildLeafNodeProgressPayload(event *dashboardevents.LeafNodeProgress) ([]byte, error) {
	payload := ExecutionPayload{
		Action:        "leaf_node_progress",
		DashboardNode: event.Node,
	}
	return json.Marshal(payload)
}

func buildLeafNodeCompletePayload(event *dashboardevents.LeafNodeComplete) ([]byte, error) {
	payload := ExecutionPayload{
		Action:        "leaf_node_complete",
		DashboardNode: event.Node,
	}
	return json.Marshal(payload)
}

func buildExecutionStartedPayload(event *dashboardevents.ExecutionStarted) ([]byte, error) {
	payload := ExecutionPayload{
		Action:        "execution_started",
		DashboardNode: event.DashboardNode,
	}
	a, _ := json.MarshalIndent(payload, "", " ")
	b := string(a)
	fmt.Println(b)
	return json.Marshal(payload)
}

func buildExecutionCompletePayload(event *dashboardevents.ExecutionComplete) ([]byte, error) {
	payload := ExecutionPayload{
		Action:        "execution_complete",
		DashboardNode: event.Dashboard,
	}
	return json.Marshal(payload)
}

func getDashboardsInterestedInResourceChanges(dashboardsBeingWatched []string, existingChangedDashboardNames []string, changedItems []*modconfig.DashboardTreeItemDiffs) []string {
	var changedDashboardNames []string

	for _, changedItem := range changedItems {
		paths := changedItem.Item.GetPaths()
		for _, nodePath := range paths {
			for _, nodeName := range nodePath {
				resourceParts, _ := modconfig.ParseResourceName(nodeName)
				// We only care about changes from these resource types
				if resourceParts.ItemType != modconfig.BlockTypeDashboard {
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

// Start starts the API server
func (s *Server) Start() {
	go s.Init(s.context)
	go StartAPI(s.context, s.webSocket)
}

// Shutdown stops the API server
func (s *Server) Shutdown(ctx context.Context) {
	// Close the DB client
	if s.dbClient != nil {
		s.dbClient.Close(ctx)
	}

	if s.webSocket != nil {
		s.webSocket.Close()
	}

	// Close the workspace
	if s.workspace != nil {
		s.workspace.Close()
	}
}

func (s *Server) HandleWorkspaceUpdate(event dashboardevents.DashboardEvent) {
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
		log.Println("[TRACE] Got workspace error event", *e)
		payload, payloadError = buildWorkspaceErrorPayload(e)
		if payloadError != nil {
			return
		}
		s.webSocket.Broadcast(payload)
		outputError(s.context, e.Error)

	case *dashboardevents.ExecutionStarted:
		log.Println("[TRACE] Got execution started event", *e)
		payload, payloadError = buildExecutionStartedPayload(e)
		if payloadError != nil {
			return
		}
		s.mutex.Lock()
		s.writePayloadToSession(e.Session, payload)
		s.mutex.Unlock()
		outputWait(s.context, fmt.Sprintf("Dashboard execution started: %s", e.DashboardNode.GetName()))

	case *dashboardevents.LeafNodeError:
		log.Println("[TRACE] Got leaf node error event", *e)

	case *dashboardevents.LeafNodeProgress:
		log.Println("[TRACE] Got leaf node complete event", *e)
		payload, payloadError = buildLeafNodeProgressPayload(e)
		if payloadError != nil {
			return
		}
		s.mutex.Lock()
		s.writePayloadToSession(e.Session, payload)
		s.mutex.Unlock()

	case *dashboardevents.LeafNodeComplete:
		log.Println("[TRACE] Got leaf node complete event", *e)
		payload, payloadError = buildLeafNodeCompletePayload(e)
		if payloadError != nil {
			return
		}
		s.mutex.Lock()
		s.writePayloadToSession(e.Session, payload)
		s.mutex.Unlock()

	case *dashboardevents.DashboardChanged:
		log.Println("[TRACE] Got dashboard changed event", *e)
		deletedDashboards := e.DeletedDashboards
		newDashboards := e.NewDashboards

		changedContainers := e.ChangedContainers
		changedBenchmarks := e.ChangedBenchmarks
		changedControls := e.ChangedControls
		changedCards := e.ChangedCards
		changedCharts := e.ChangedCharts
		changedHierarchies := e.ChangedHierarchies
		changedImages := e.ChangedImages
		changedInputs := e.ChangedInputs
		changedTables := e.ChangedTables
		changedTexts := e.ChangedTexts
		changedDashboards := e.ChangedDashboards

		// If nothing has changed, ignore
		if len(deletedDashboards) == 0 &&
			len(newDashboards) == 0 &&
			len(changedContainers) == 0 &&
			len(changedBenchmarks) == 0 &&
			len(changedControls) == 0 &&
			len(changedCards) == 0 &&
			len(changedCharts) == 0 &&
			len(changedHierarchies) == 0 &&
			len(changedImages) == 0 &&
			len(changedInputs) == 0 &&
			len(changedTables) == 0 &&
			len(changedTexts) == 0 &&
			len(changedDashboards) == 0 {
			return
		}

		for k, v := range s.dashboardClients {
			log.Printf("[TRACE] Dashboard client: %v %v\n", k, typeHelpers.SafeString(v.Dashboard))
		}

		// If) any deleted/new/changed dashboards, emit an available dashboards message to clients
		if len(deletedDashboards) != 0 || len(newDashboards) != 0 || len(changedDashboards) != 0 {
			outputMessage(s.context, "Available Dashboards updated")
			payload, payloadError = buildAvailableDashboardsPayload(s.workspace.GetResourceMaps())
			if payloadError != nil {
				return
			}
			s.webSocket.Broadcast(payload)
		}

		var dashboardssBeingWatched []string
		s.mutex.Lock()
		for _, dashboardClientInfo := range s.dashboardClients {
			dashboardName := typeHelpers.SafeString(dashboardClientInfo.Dashboard)
			if dashboardClientInfo.Dashboard != nil {
				if helpers.StringSliceContains(dashboardssBeingWatched, dashboardName) {
					continue
				}
				dashboardssBeingWatched = append(dashboardssBeingWatched, dashboardName)
			}
		}
		s.mutex.Unlock()

		var changedDashboardNames []string
		var newDashboardNames []string

		// Process the changed items and make a note of the dashboard(s) they're in
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedContainers)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedBenchmarks)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedControls)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedCards)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedCharts)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedHierarchies)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedImages)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedInputs)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedTables)...)
		changedDashboardNames = append(changedDashboardNames, getDashboardsInterestedInResourceChanges(dashboardssBeingWatched, changedDashboardNames, changedTexts)...)

		for _, changedDashboard := range changedDashboards {
			if helpers.StringSliceContains(changedDashboardNames, changedDashboard.Name) {
				continue
			}
			changedDashboardNames = append(changedDashboardNames, changedDashboard.Name)
		}

		for _, changedDashboardName := range changedDashboardNames {
			s.mutex.Lock()
			for sessionId, dashboardClientInfo := range s.dashboardClients {
				if typeHelpers.SafeString(dashboardClientInfo.Dashboard) == changedDashboardName {
					dashboardexecute.Executor.ExecuteDashboard(s.context, sessionId, changedDashboardName, nil, s.workspace, s.dbClient)
				}
			}
			s.mutex.Unlock()
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

		for _, newDashboardName := range newDashboardNames {
			s.mutex.Lock()
			for sessionId, dashboardClientInfo := range s.dashboardClients {
				if typeHelpers.SafeString(dashboardClientInfo.Dashboard) == newDashboardName {
					dashboardexecute.Executor.ExecuteDashboard(s.context, sessionId, newDashboardName, nil, s.workspace, s.dbClient)
				}
			}
			s.mutex.Unlock()
		}

	case *dashboardevents.DashboardError:
		log.Println("[TRACE] Got dashboard error event", *e)

	case *dashboardevents.DashboardComplete:
		log.Println("[TRACE] Got dashboard complete event", *e)

	case *dashboardevents.ExecutionComplete:
		log.Println("[TRACE] Got execution complete event", *e)
		payload, payloadError = buildExecutionCompletePayload(e)
		if payloadError != nil {
			return
		}
		dashboardName := e.Dashboard.GetName()
		s.mutex.Lock()
		s.writePayloadToSession(e.Session, payload)
		s.mutex.Unlock()
		outputReady(s.context, fmt.Sprintf("Execution complete: %s", dashboardName))
	}
}

func (s *Server) Init(ctx context.Context) {
	// Return list of dashboards on connect
	s.webSocket.HandleConnect(func(session *melody.Session) {
		log.Println("[TRACE] Client connected")
		s.addSession(session)
	})

	s.webSocket.HandleDisconnect(func(session *melody.Session) {
		log.Println("[TRACE] Client disconnected")
		s.clearSession(ctx, session)
	})

	s.webSocket.HandleMessage(s.handleMessageFunc(ctx))
	outputMessage(ctx, "Initialization complete")
}

func (s *Server) handleMessageFunc(ctx context.Context) func(session *melody.Session, msg []byte) {
	return func(session *melody.Session, msg []byte) {
		log.Println("[TRACE] Got message", string(msg))

		// TODO TEMP GET SESSION ID
		sessionId := s.getSessionId(session)

		var request ClientRequest
		// if we could not decode message - ignore
		if err := json.Unmarshal(msg, &request); err == nil {
			switch request.Action {
			case "get_dashboard_metadata":
				payload, err := buildDashboardMetadataPayload(s.workspace.GetResourceMaps())
				if err != nil {
					panic(fmt.Errorf("error building payload for get_metadata: %v", err))
				}
				session.Write(payload)
			case "get_available_dashboards":
				payload, err := buildAvailableDashboardsPayload(s.workspace.GetResourceMaps())
				if err != nil {
					panic(fmt.Errorf("error building payload for get_available_dashboards: %v", err))
				}
				session.Write(payload)
			case "select_dashboard":
				s.setDashboardForSession(sessionId, request.Payload.Dashboard.FullName)
				dashboardexecute.Executor.ExecuteDashboard(ctx, sessionId, request.Payload.Dashboard.FullName, request.Payload.InputValues, s.workspace, s.dbClient)
			case "set_dashboard_inputs":
				dashboardexecute.Executor.SetDashboardInputs(ctx, sessionId, request.Payload.InputValues)
			case "clear_dashboard":
				dashboardexecute.Executor.ClearDashboard(ctx, sessionId)
			}
		}
	}
}

func (s *Server) setDashboardForSession(sessionID string, dashboardName string) *DashboardClientInfo {
	s.mutex.Lock()
	dashboardClientInfo := s.dashboardClients[sessionID]
	dashboardClientInfo.Dashboard = &dashboardName
	s.mutex.Unlock()
	return dashboardClientInfo
}
func (s *Server) clearSession(ctx context.Context, session *melody.Session) {
	s.mutex.Lock()
	sessionId := s.getSessionId(session)
	dashboardexecute.Executor.ClearDashboard(ctx, sessionId)
	delete(s.dashboardClients, sessionId)
	s.mutex.Unlock()
}

func (s *Server) addSession(session *melody.Session) {
	s.mutex.Lock()
	sessionId := s.getSessionId(session)
	s.dashboardClients[sessionId] = &DashboardClientInfo{
		Session: session,
	}
	s.mutex.Unlock()
}

func (s *Server) getSessionId(session *melody.Session) string {
	return fmt.Sprintf("%p", session)
}

func (s *Server) writePayloadToSession(sessionID string, payload []byte) {
	if sessionInfo, ok := s.dashboardClients[sessionID]; ok {
		sessionInfo.Session.Write(payload)
	}
}
