package dashboardserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
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
	context            context.Context
	dbClient           db_common.Client
	mutex              *sync.Mutex
	dashboardClients   map[*melody.Session]*DashboardClientInfo
	webSocket          *melody.Melody
	workspace          *workspace.Workspace
	workspaceResources *modconfig.WorkspaceResourceMaps
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
	Dashboard *string
}

func NewServer(ctx context.Context, dbClient db_common.Client) (*Server, error) {
	outputWait(ctx, "Starting Dashboard Server")
	loadedWorkspace, err := workspace.Load(ctx, viper.GetString(constants.ArgWorkspaceChDir))
	if err != nil {
		return nil, err
	}

	webSocket := melody.New()

	var dashboardClients = make(map[*melody.Session]*DashboardClientInfo)

	var mutex = &sync.Mutex{}

	server := &Server{
		context:            ctx,
		dbClient:           dbClient,
		mutex:              mutex,
		dashboardClients:   dashboardClients,
		webSocket:          webSocket,
		workspace:          loadedWorkspace,
		workspaceResources: loadedWorkspace.GetResourceMaps(),
	}

	loadedWorkspace.RegisterDashboardEventHandler(server.HandleWorkspaceUpdate)
	err = loadedWorkspace.SetupWatcher(ctx, dbClient, func(c context.Context, e error) {})
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

func buildAvailableDashboardsPayload(workspace *modconfig.WorkspaceResourceMaps) ([]byte, error) {
	dashboardsByMod := make(map[string]map[string]ModAvailableDashboard)
	for _, mod := range workspace.Mods {
		_, ok := dashboardsByMod[mod.FullName]
		if !ok {
			dashboardsByMod[mod.FullName] = make(map[string]ModAvailableDashboard)
		}
		for _, dashboard := range mod.Dashboards {
			dashboardsByMod[mod.FullName][dashboard.FullName] = ModAvailableDashboard{
				Title:     typeHelpers.SafeString(dashboard.Title),
				FullName:  dashboard.FullName,
				ShortName: dashboard.ShortName,
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
	//jsonString, _ := json.Marshal(payload)
	//return jsonString
	jsonString, err := json.MarshalIndent(payload, "", "  ")
	fmt.Println(err)
	a := string(jsonString)
	fmt.Println(a)

	return jsonString, err
}

func buildExecutionStartedPayload(event *dashboardevents.ExecutionStarted) ([]byte, error) {
	payload := ExecutionPayload{
		Action:        "execution_started",
		DashboardNode: event.DashboardNode,
	}
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

	// reload workspace
	s.workspaceResources = s.workspace.GetResourceMaps()

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
		dashboardName := e.DashboardNode.GetName()
		s.mutex.Lock()
		for session, repoInfo := range s.dashboardClients {
			// If this session is interested in this dashboard, broadcast to it
			if (repoInfo.Dashboard != nil) && *repoInfo.Dashboard == dashboardName {
				session.Write(payload)
			}
		}
		s.mutex.Unlock()
		outputWait(s.context, fmt.Sprintf("Dashboard execution started: %s", dashboardName))

	case *dashboardevents.LeafNodeError:
		log.Println("[TRACE] Got leaf node error event", *e)

	case *dashboardevents.LeafNodeProgress:
		log.Println("[TRACE] Got leaf node complete event", *e)
		payload, payloadError = buildLeafNodeProgressPayload(e)
		if payloadError != nil {
			return
		}
		paths := e.Node.GetPath()
		s.mutex.Lock()
		for session, repoInfo := range s.dashboardClients {
			// If this session is interested in this dashboard, broadcast to it
			if (repoInfo.Dashboard != nil) && helpers.StringSliceContains(paths, *repoInfo.Dashboard) {
				session.Write(payload)
			}
		}
		s.mutex.Unlock()

	case *dashboardevents.LeafNodeComplete:
		log.Println("[TRACE] Got leaf node complete event", *e)
		payload, payloadError = buildLeafNodeCompletePayload(e)
		if payloadError != nil {
			return
		}
		paths := e.Node.GetPath()
		s.mutex.Lock()
		for session, repoInfo := range s.dashboardClients {
			// If this session is interested in this dashboard, broadcast to it
			if (repoInfo.Dashboard != nil) && helpers.StringSliceContains(paths, *repoInfo.Dashboard) {
				session.Write(payload)
			}
		}
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
			payload, payloadError = buildAvailableDashboardsPayload(s.workspaceResources)
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
			if helpers.StringSliceContains(dashboardssBeingWatched, changedDashboardName) {
				dashboardexecute.ExecuteDashboardNode(s.context, changedDashboardName, s.workspace, s.dbClient)
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

		for _, newDashboardName := range newDashboardNames {
			if helpers.StringSliceContains(dashboardssBeingWatched, newDashboardName) {
				dashboardexecute.ExecuteDashboardNode(s.context, newDashboardName, s.workspace, s.dbClient)
			}
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
		for session, repoInfo := range s.dashboardClients {
			// If this session is interested in this dashboard, broadcast to it
			if (repoInfo.Dashboard != nil) && *repoInfo.Dashboard == dashboardName {
				session.Write(payload)
			}
		}
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
		s.clearSession(session)
	})

	s.webSocket.HandleMessage(func(session *melody.Session, msg []byte) {
		log.Println("[TRACE] Got message", string(msg))
		var request ClientRequest
		// if we could not decode message - ignore
		if err := json.Unmarshal(msg, &request); err == nil {

			switch request.Action {
			case "get_dashboard_metadata":
				payload, err := buildDashboardMetadataPayload(s.workspaceResources)
				if err != nil {
					panic(fmt.Errorf("error building payload for get_metadata: %v", err))
				}
				session.Write(payload)
			case "get_available_dashboards":
				payload, err := buildAvailableDashboardsPayload(s.workspaceResources)
				if err != nil {
					panic(fmt.Errorf("error building payload for get_available_dashboards: %v", err))
				}
				session.Write(payload)
			case "select_dashboard":
				log.Printf("[TRACE] Got event: %v\n", request.Payload.Dashboard)
				dashboardClientInfo := s.getSession(session)
				dashboardClientInfo.Dashboard = &request.Payload.Dashboard.FullName
				dashboardexecute.ExecuteDashboardNode(ctx, request.Payload.Dashboard.FullName, s.workspace, s.dbClient)
			}
		}
	})
	outputMessage(ctx, "Initialization complete")
}

func (s *Server) getSession(session *melody.Session) *DashboardClientInfo {
	s.mutex.Lock()
	dashboardClientInfo := s.dashboardClients[session]
	s.mutex.Unlock()
	return dashboardClientInfo
}

func (s *Server) clearSession(session *melody.Session) {
	s.mutex.Lock()
	delete(s.dashboardClients, session)
	s.mutex.Unlock()
}

func (s *Server) addSession(session *melody.Session) {
	s.mutex.Lock()
	s.dashboardClients[session] = &DashboardClientInfo{}
	s.mutex.Unlock()
}
