package reportserver

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/go-kit/types"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/db/db_local"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportexecute"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
	"gopkg.in/olahol/melody.v1"
)

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

func NewServer(ctx context.Context) (*Server, error) {
	var dbClient, err = db_local.GetLocalClient(ctx, constants.InvokerReport)
	if err != nil {
		return nil, err
	}

	refreshResult := dbClient.RefreshConnectionAndSearchPaths(ctx)
	if err != nil {
		return nil, err
	}
	refreshResult.ShowWarnings()

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
	err = loadedWorkspace.SetupWatcher(ctx, dbClient, nil)

	return server, err
}

func buildAvailableReportsPayload(reports map[string]*modconfig.ReportContainer) []byte {
	reportsPayload := make(map[string]string)
	for _, report := range reports {
		reportsPayload[report.FullName] = types.SafeString(report.Title)
	}
	payload := AvailableReportsPayload{
		Action:  "available_reports",
		Reports: reportsPayload,
	}
	jsonString, _ := json.Marshal(payload)
	return jsonString
}

func buildWorkspaceErrorPayload(e *reportevents.WorkspaceError) []byte {
	payload := ErrorPayload{
		Action: "workspace_error",
		Error:  e.Error.Error(),
	}
	jsonString, _ := json.Marshal(payload)
	return jsonString
}

func buildLeafNodeCompletePayload(event *reportevents.LeafNodeComplete) []byte {
	payload := ExecutionPayload{
		Action:     "leaf_node_complete",
		ReportNode: event.Node,
	}
	jsonString, _ := json.Marshal(payload)
	return jsonString
}

func buildExecutionStartedPayload(event *reportevents.ExecutionStarted) []byte {
	payload := ExecutionPayload{
		Action:     "execution_started",
		ReportNode: event.ReportNode,
	}
	jsonString, _ := json.Marshal(payload)
	return jsonString
}

func buildExecutionCompletePayload(event *reportevents.ExecutionComplete) []byte {
	payload := ExecutionPayload{
		Action:     "execution_complete",
		ReportNode: event.Report,
	}
	jsonString, err := json.MarshalIndent(payload, "", "  ")
	fmt.Println(err)
	a := string(jsonString)
	fmt.Println(a)
	return jsonString
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

// Starts the API server
func (s *Server) Start() {
	go Init(s.context, s.webSocket, s.workspace, s.dbClient, s.reportClients, s.mutex)
	StartAPI(s.context, s.webSocket)
}

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
	// Possible events - TODO work out best way to handle these
	/*
		WORKSPACE_ERROR
		EXECUTION_STARTED
		COUNTER_CHANGED
		COUNTER_ERROR
		COUNTER_COMPLETE
		REPORT_CHANGED
		REPORT_ERROR
		REPORT_COMPLETE
		EXECUTION_COMPLETE
	*/

	fmt.Println("Got workspace update event", event)
	switch e := event.(type) {

	case *reportevents.WorkspaceError:
		fmt.Println("Got workspace error event", *e)
		payload := buildWorkspaceErrorPayload(e)
		s.webSocket.Broadcast(payload)

	case *reportevents.ExecutionStarted:
		fmt.Println("Got execution started event", *e)
		payload := buildExecutionStartedPayload(e)
		reportName := e.ReportNode.GetName()
		s.mutex.Lock()
		for session, repoInfo := range s.reportClients {
			// If this session is interested in this report, broadcast to it
			if (repoInfo.Report != nil) && *repoInfo.Report == reportName {
				session.Write(payload)
			}
		}
		s.mutex.Unlock()

	case *reportevents.LeafNodeError:
		fmt.Println("Got leaf node error event", *e)

	case *reportevents.LeafNodeComplete:
		fmt.Println("Got leaf node complete event", *e)
		payload := buildLeafNodeCompletePayload(e)
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
		fmt.Println("Got report changed event", *e)
		deletedReports := e.DeletedReports
		newReports := e.NewReports

		changedContainers := e.ChangedContainers
		changedBenchmarks := e.ChangedBenchmarks
		changedControls := e.ChangedControls
		changedCharts := e.ChangedCharts
		changedCounters := e.ChangedCounters
		changedHierarchies := e.ChangedHierarchies
		changedImages := e.ChangedImages
		changedTables := e.ChangedTables
		changedTexts := e.ChangedTexts
		changedReports := e.ChangedReports

		// If nothing has changed, ignore
		if len(deletedReports) == 0 &&
			len(newReports) == 0 &&
			len(changedContainers) == 0 &&
			len(changedBenchmarks) == 0 &&
			len(changedControls) == 0 &&
			len(changedCharts) == 0 &&
			len(changedCounters) == 0 &&
			len(changedHierarchies) == 0 &&
			len(changedImages) == 0 &&
			len(changedTables) == 0 &&
			len(changedTexts) == 0 &&
			len(changedReports) == 0 {
			return
		}

		for k, v := range s.reportClients {
			fmt.Printf("Report client: %v %v\n", k, types.SafeString(v.Report))
		}

		// If) any deleted/new/changed reports, emit an available reports message to clients
		if len(deletedReports) != 0 || len(newReports) != 0 || len(changedReports) != 0 {
			s.webSocket.Broadcast(buildAvailableReportsPayload(s.workspace.Mod.Reports))
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
		changedReportNames = append(changedReportNames, getReportsInterestedInResourceChanges(reportsBeingWatched, changedReportNames, changedCharts)...)
		changedReportNames = append(changedReportNames, getReportsInterestedInResourceChanges(reportsBeingWatched, changedReportNames, changedCounters)...)
		changedReportNames = append(changedReportNames, getReportsInterestedInResourceChanges(reportsBeingWatched, changedReportNames, changedHierarchies)...)
		changedReportNames = append(changedReportNames, getReportsInterestedInResourceChanges(reportsBeingWatched, changedReportNames, changedImages)...)
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
		fmt.Println("Got report error event", *e)

	case *reportevents.ReportComplete:
		fmt.Println("Got report complete event", *e)

	case *reportevents.ExecutionComplete:
		fmt.Println("Got execution complete event", *e)
		payload := buildExecutionCompletePayload(e)
		reportName := e.Report.GetName()
		s.mutex.Lock()
		for session, repoInfo := range s.reportClients {
			// If this session is interested in this report, broadcast to it
			if (repoInfo.Report != nil) && *repoInfo.Report == reportName {
				session.Write(payload)
			}
		}
		s.mutex.Unlock()
	}
}
