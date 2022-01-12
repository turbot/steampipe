package reportserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"github.com/turbot/go-kit/helpers"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/db/db_local"
	"gopkg.in/olahol/melody.v1"

	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/executionlayer"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportinterfaces"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
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

func buildPanelCompletePayload(event *reportevents.PanelComplete) []byte {
	payload := ExecutionPayload{
		Action:     "panel_complete",
		ReportNode: event.Panel,
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
	jsonString, _ := json.Marshal(payload)
	return jsonString
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
		PANEL_CHANGED
		PANEL_ERROR
		PANEL_COMPLETE
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

	case *reportevents.PanelError:
		fmt.Println("Got panel error event", *e)

	case *reportevents.PanelComplete:
		fmt.Println("Got panel complete event", *e)
		payload := buildPanelCompletePayload(e)
		panelName := e.Panel.GetName()
		s.mutex.Lock()
		for session, repoInfo := range s.reportClients {
			// If this session is interested in this report, broadcast to it
			if (repoInfo.Report != nil) && strings.HasPrefix(panelName, *repoInfo.Report) {
				session.Write(payload)
			}
		}
		s.mutex.Unlock()

	case *reportevents.ReportChanged:
		fmt.Println("Got report changed event", *e)
		deletedReports := e.DeletedReports
		newReports := e.NewReports
		changedPanels := e.ChangedPanels
		changedReports := e.ChangedReports

		// If nothing has changed, ignore
		if len(deletedReports) == 0 && len(newReports) == 0 && len(changedPanels) == 0 && len(changedReports) == 0 {
			return
		}

		for k, v := range s.reportClients {
			fmt.Printf("Report client: %v %v\n", k, types.SafeString(v.Report))
		}

		// If) any deleted/new/changed reports, emit an available reports message to clients
		if len(deletedReports) != 0 || len(newReports) != 0 || len(changedReports) != 0 {
			s.webSocket.Broadcast(buildAvailableReportsPayload(s.workspace.Mod.Reports))
		}

		// If we have no changed panels or reports, ignore the message for now
		if len(deletedReports) == 0 && len(newReports) == 0 && len(changedPanels) == 0 && len(changedReports) == 0 {
			return
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

		// Capture the changed panels and make a note of the report(s) they're in
		for _, changedPanel := range changedPanels {
			paths := changedPanel.Item.GetPaths()
			for _, nodePath := range paths {
				for _, nodeName := range nodePath {
					resourceParts, _ := modconfig.ParseResourceName(nodeName)
					if resourceParts.ItemType != modconfig.BlockTypeReport {
						continue
					}
					if helpers.StringSliceContains(changedReportNames, nodeName) {
						continue
					}
					changedReportNames = append(changedReportNames, nodeName)
				}
			}
		}

		for _, changedReport := range changedReports {
			if helpers.StringSliceContains(changedReportNames, changedReport.Name) {
				continue
			}
			changedReportNames = append(changedReportNames, changedReport.Name)
		}

		for _, changedReportName := range changedReportNames {
			if helpers.StringSliceContains(reportsBeingWatched, changedReportName) {
				executionlayer.ExecuteReportNode(s.context, changedReportName, s.workspace, s.dbClient)
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
				executionlayer.ExecuteReportNode(s.context, newReportName, s.workspace, s.dbClient)
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
