package reportserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/turbot/go-kit/helpers"
	typeHelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe/executionlayer"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"sync"

	"github.com/spf13/viper"
	"gopkg.in/olahol/melody.v1"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportexecute"
	"github.com/turbot/steampipe/workspace"
)

type Server struct {
	context       context.Context
	dbClient      *db.Client
	mutex         *sync.Mutex
	reportClients map[*melody.Session]*ReportClientInfo
	webSocket     *melody.Melody
	workspace     *workspace.Workspace
}

type ExecutionPayload struct {
	Action string                   `json:"action"`
	Report *reportexecute.ReportRun `json:"report"`
}

type ReportClientInfo struct {
	Report *string
}

func NewServer(ctx context.Context) (*Server, error) {
	dbClient, err := db.NewClient(true)
	if err != nil {
		return nil, err
	}

	loadedWorkspace, err := workspace.Load(viper.GetString(constants.ArgWorkspace))
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
	err = loadedWorkspace.SetupWatcher(dbClient)

	return server, err
}

func buildExecutionStartedPayload(event *reportevents.ExecutionStarted) []byte {
	payload := ExecutionPayload{
		Action: "execution_started",
		Report: event.Report,
	}
	jsonString, _ := json.Marshal(payload)
	return jsonString
}

func buildExecutionCompletePayload(event *reportevents.ExecutionComplete) []byte {
	payload := ExecutionPayload{
		Action: "execution_complete",
		Report: event.Report,
	}
	jsonString, _ := json.Marshal(payload)
	return jsonString
}

// Starts the API server
func (s *Server) Start() {
	go Init(s.context, s.webSocket, s.workspace, s.dbClient, s.reportClients, s.mutex)
	StartAPI(s.context, s.webSocket)
}

func (s *Server) Shutdown() {
	// Close the DB client
	if s.dbClient != nil {
		s.dbClient.Close()
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
	// TODO ...
	fmt.Println("Got update event", event)
	switch e := event.(type) {
	case *reportevents.ExecutionStarted:
		fmt.Println("Got execution started event", *e)
		payload := buildExecutionStartedPayload(e)
		reportName := e.Report.Name
		s.mutex.Lock()
		for session, repoInfo := range s.reportClients {
			// If this session is interested in this report, broadcast to it
			if (repoInfo.Report != nil) && *repoInfo.Report == reportName {
				session.Write(payload)
			}
		}
		s.mutex.Unlock()

	case *reportevents.ExecutionComplete:
		fmt.Println("Got execution complete event", *e)
		payload := buildExecutionCompletePayload(e)
		reportName := e.Report.Name
		s.mutex.Lock()
		for session, repoInfo := range s.reportClients {
			// If this session is interested in this report, broadcast to it
			if (repoInfo.Report != nil) && *repoInfo.Report == reportName {
				session.Write(payload)
			}
		}
		s.mutex.Unlock()
	case *reportevents.ReportChanged:
		fmt.Println("Got report changed event", *e)
		changedPanels := e.ChangedPanels
		changedReports := e.ChangedReports

		// If we have no changed panels or reports, ignore the message for now
		if len(changedPanels) == 0 && len(changedReports) == 0 {
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
				executionlayer.ExecuteReport(s.context, changedReportName, s.workspace, s.dbClient)
			}
		}
	}
}
