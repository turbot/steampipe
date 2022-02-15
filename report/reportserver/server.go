package reportserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/spf13/viper"
	"gopkg.in/olahol/melody.v1"

	"github.com/turbot/go-kit/helpers"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportexecute"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
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
	context       context.Context
	dbClient      db_common.Client
	mutex         *sync.Mutex
	reportClients map[*melody.Session]*ReportClientInfo
	webSocket     *melody.Melody
	workspace     *workspace.Workspace
}

type ErrorPayload struct {
	Action string `json:"action"`
	Error  string `json:"error"`
}

type ExecutionPayload struct {
	Action     string                         `json:"action"`
	ReportNode reportinterfaces.ReportNodeRun `json:"report_node"`
}

type ReportClientInfo struct {
	Report *string
}

func NewServer(ctx context.Context, dbClient db_common.Client) (*Server, error) {
	outputWait(ctx, "Starting Report Server")
	loadedWorkspace, err := workspace.Load(ctx, viper.GetString(constants.ArgWorkspaceChDir))
	if err != nil {
		return nil, err
	}

	webSocket := melody.New()

	var reportClients = make(map[*melody.Session]*ReportClientInfo)

	var mutex = &sync.Mutex{}

	server := &Server{
		context:       ctx,
		dbClient:      dbClient,
		mutex:         mutex,
		reportClients: reportClients,
		webSocket:     webSocket,
		workspace:     loadedWorkspace,
	}

	loadedWorkspace.RegisterReportEventHandler(server.HandleWorkspaceUpdate)
	err = loadedWorkspace.SetupWatcher(ctx, dbClient, func(c context.Context, e error) {})
	outputMessage(ctx, "Workspace loaded")

	return server, err
}

func buildDashboardMetadataPayload(workspace *workspace.Workspace) ([]byte, error) {
	installedMods := make(map[string]ModDashboardMetadata)
	for _, mod := range workspace.Mods {
		// Ignore current mod
		if mod.FullName == workspace.Mod.FullName {
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
				Title:     typeHelpers.SafeString(workspace.Mod.Title),
				FullName:  workspace.Mod.FullName,
				ShortName: workspace.Mod.ShortName,
			},
			InstalledMods: installedMods,
		},
	}
	return json.Marshal(payload)
}

func buildAvailableDashboardsPayload(workspace *workspace.Workspace) ([]byte, error) {
	dashboardsByMod := make(map[string]map[string]ModAvailableDashboard)
	for _, mod := range workspace.Mods {
		_, ok := dashboardsByMod[mod.FullName]
		if !ok {
			dashboardsByMod[mod.FullName] = make(map[string]ModAvailableDashboard)
		}
		for _, report := range mod.Reports {
			dashboardsByMod[mod.FullName][report.FullName] = ModAvailableDashboard{
				Title:     typeHelpers.SafeString(report.Title),
				FullName:  report.FullName,
				ShortName: report.ShortName,
			}
		}
	}
	payload := AvailableDashboardsPayload{
		Action:          "available_dashboards",
		DashboardsByMod: dashboardsByMod,
	}
	return json.Marshal(payload)
}

func buildWorkspaceErrorPayload(e *reportevents.WorkspaceError) ([]byte, error) {
	payload := ErrorPayload{
		Action: "workspace_error",
		Error:  e.Error.Error(),
	}
	return json.Marshal(payload)
}

func buildLeafNodeProgressPayload(event *reportevents.LeafNodeProgress) ([]byte, error) {
	payload := ExecutionPayload{
		Action:     "leaf_node_progress",
		ReportNode: event.Node,
	}
	return json.Marshal(payload)
}

func buildLeafNodeCompletePayload(event *reportevents.LeafNodeComplete) ([]byte, error) {
	payload := ExecutionPayload{
		Action:     "leaf_node_complete",
		ReportNode: event.Node,
	}
	//jsonString, _ := json.Marshal(payload)
	//return jsonString
	jsonString, err := json.MarshalIndent(payload, "", "  ")
	fmt.Println(err)
	a := string(jsonString)
	fmt.Println(a)

	return jsonString, err
}

func buildExecutionStartedPayload(event *reportevents.ExecutionStarted) ([]byte, error) {
	payload := ExecutionPayload{
		Action:     "execution_started",
		ReportNode: event.ReportNode,
	}
	return json.Marshal(payload)
}

func buildExecutionCompletePayload(event *reportevents.ExecutionComplete) ([]byte, error) {
	payload := ExecutionPayload{
		Action:     "execution_complete",
		ReportNode: event.Report,
	}
	return json.Marshal(payload)
}

func getReportsInterestedInResourceChanges(reportsBeingWatched []string, existingChangedReportNames []string, changedItems []*modconfig.ReportTreeItemDiffs) []string {
	var changedReportNames []string

	for _, changedItem := range changedItems {
		paths := changedItem.Item.GetPaths()
		for _, nodePath := range paths {
			for _, nodeName := range nodePath {
				resourceParts, _ := modconfig.ParseResourceName(nodeName)
				// We only care about changes from these resource types
				if resourceParts.ItemType != modconfig.BlockTypeReport {
					continue
				}

				if helpers.StringSliceContains(existingChangedReportNames, nodeName) || helpers.StringSliceContains(changedReportNames, nodeName) || !helpers.StringSliceContains(reportsBeingWatched, nodeName) {
					continue
				}

				changedReportNames = append(changedReportNames, nodeName)
			}
		}
	}

	return changedReportNames
}

// Start starts the API server
func (s *Server) Start() {
	go Init(s.context, s.webSocket, s.workspace, s.dbClient, s.reportClients, s.mutex)
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

func (s *Server) HandleWorkspaceUpdate(event reportevents.ReportEvent) {
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

	case *reportevents.WorkspaceError:
		log.Println("[TRACE] Got workspace error event", *e)
		payload, payloadError = buildWorkspaceErrorPayload(e)
		if payloadError != nil {
			return
		}
		s.webSocket.Broadcast(payload)
		outputError(s.context, e.Error)

	case *reportevents.ExecutionStarted:
		log.Println("[TRACE] Got execution started event", *e)
		payload, payloadError = buildExecutionStartedPayload(e)
		if payloadError != nil {
			return
		}
		reportName := e.ReportNode.GetName()
		s.mutex.Lock()
		for session, repoInfo := range s.reportClients {
			// If this session is interested in this report, broadcast to it
			if (repoInfo.Report != nil) && *repoInfo.Report == reportName {
				session.Write(payload)
			}
		}
		s.mutex.Unlock()
		outputWait(s.context, fmt.Sprintf("Report execution started: %s", reportName))

	case *reportevents.LeafNodeError:
		log.Println("[TRACE] Got leaf node error event", *e)

	case *reportevents.LeafNodeProgress:
		log.Println("[TRACE] Got leaf node complete event", *e)
		payload, payloadError = buildLeafNodeProgressPayload(e)
		if payloadError != nil {
			return
		}
		paths := e.Node.GetPath()
		s.mutex.Lock()
		for session, repoInfo := range s.reportClients {
			// If this session is interested in this report, broadcast to it
			if (repoInfo.Report != nil) && helpers.StringSliceContains(paths, *repoInfo.Report) {
				session.Write(payload)
			}
		}
		s.mutex.Unlock()

	case *reportevents.LeafNodeComplete:
		log.Println("[TRACE] Got leaf node complete event", *e)
		payload, payloadError = buildLeafNodeCompletePayload(e)
		if payloadError != nil {
			return
		}
		paths := e.Node.GetPath()
		s.mutex.Lock()
		for session, repoInfo := range s.reportClients {
			// If this session is interested in this report, broadcast to it
			if (repoInfo.Report != nil) && helpers.StringSliceContains(paths, *repoInfo.Report) {
				session.Write(payload)
			}
		}
		s.mutex.Unlock()

	case *reportevents.ReportChanged:
		log.Println("[TRACE] Got report changed event", *e)
		deletedReports := e.DeletedReports
		newReports := e.NewReports

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
		changedReports := e.ChangedReports

		// If nothing has changed, ignore
		if len(deletedReports) == 0 &&
			len(newReports) == 0 &&
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
			len(changedReports) == 0 {
			return
		}

		for k, v := range s.reportClients {
			log.Printf("[TRACE] Report client: %v %v\n", k, typeHelpers.SafeString(v.Report))
		}

		// If) any deleted/new/changed reports, emit an available reports message to clients
		if len(deletedReports) != 0 || len(newReports) != 0 || len(changedReports) != 0 {
			outputMessage(s.context, "Available Reports updated")
			payload, payloadError = buildAvailableDashboardsPayload(s.workspace)
			if payloadError != nil {
				return
			}
			s.webSocket.Broadcast(payload)
		}

		var reportsBeingWatched []string
		s.mutex.Lock()
		for _, reportClientInfo := range s.reportClients {
			reportName := typeHelpers.SafeString(reportClientInfo.Report)
			if reportClientInfo.Report != nil {
				if helpers.StringSliceContains(reportsBeingWatched, reportName) {
					continue
				}
				reportsBeingWatched = append(reportsBeingWatched, reportName)
			}
		}
		s.mutex.Unlock()

		var changedReportNames []string
		var newReportNames []string

		// Process the changed items and make a note of the report(s) they're in
		changedReportNames = append(changedReportNames, getReportsInterestedInResourceChanges(reportsBeingWatched, changedReportNames, changedContainers)...)
		changedReportNames = append(changedReportNames, getReportsInterestedInResourceChanges(reportsBeingWatched, changedReportNames, changedBenchmarks)...)
		changedReportNames = append(changedReportNames, getReportsInterestedInResourceChanges(reportsBeingWatched, changedReportNames, changedControls)...)
		changedReportNames = append(changedReportNames, getReportsInterestedInResourceChanges(reportsBeingWatched, changedReportNames, changedCards)...)
		changedReportNames = append(changedReportNames, getReportsInterestedInResourceChanges(reportsBeingWatched, changedReportNames, changedCharts)...)
		changedReportNames = append(changedReportNames, getReportsInterestedInResourceChanges(reportsBeingWatched, changedReportNames, changedHierarchies)...)
		changedReportNames = append(changedReportNames, getReportsInterestedInResourceChanges(reportsBeingWatched, changedReportNames, changedImages)...)
		changedReportNames = append(changedReportNames, getReportsInterestedInResourceChanges(reportsBeingWatched, changedReportNames, changedInputs)...)
		changedReportNames = append(changedReportNames, getReportsInterestedInResourceChanges(reportsBeingWatched, changedReportNames, changedTables)...)
		changedReportNames = append(changedReportNames, getReportsInterestedInResourceChanges(reportsBeingWatched, changedReportNames, changedTexts)...)

		for _, changedReport := range changedReports {
			if helpers.StringSliceContains(changedReportNames, changedReport.Name) {
				continue
			}
			changedReportNames = append(changedReportNames, changedReport.Name)
		}

		for _, changedReportName := range changedReportNames {
			if helpers.StringSliceContains(reportsBeingWatched, changedReportName) {
				reportexecute.ExecuteReportNode(s.context, changedReportName, s.workspace, s.dbClient)
			}
		}

		// Special case - if we previously had a workspace error, any previously existing reports
		// will come in here as new, so we need to check if any of those new reports are being watched.
		// If so, execute them
		for _, newReport := range newReports {
			if helpers.StringSliceContains(newReportNames, newReport.Name()) {
				continue
			}
			newReportNames = append(newReportNames, newReport.Name())
		}

		for _, newReportName := range newReportNames {
			if helpers.StringSliceContains(reportsBeingWatched, newReportName) {
				reportexecute.ExecuteReportNode(s.context, newReportName, s.workspace, s.dbClient)
			}
		}

	case *reportevents.ReportError:
		log.Println("[TRACE] Got report error event", *e)

	case *reportevents.ReportComplete:
		log.Println("[TRACE] Got report complete event", *e)

	case *reportevents.ExecutionComplete:
		log.Println("[TRACE] Got execution complete event", *e)
		payload, payloadError = buildExecutionCompletePayload(e)
		if payloadError != nil {
			return
		}
		reportName := e.Report.GetName()
		s.mutex.Lock()
		for session, repoInfo := range s.reportClients {
			// If this session is interested in this report, broadcast to it
			if (repoInfo.Report != nil) && *repoInfo.Report == reportName {
				session.Write(payload)
			}
		}
		s.mutex.Unlock()
		outputReady(s.context, fmt.Sprintf("Execution complete: %s", reportName))
	}
}
