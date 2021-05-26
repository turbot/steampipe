package reportserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"
	"gopkg.in/olahol/melody.v1"

	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/report/reporteventpublisher"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/report/reportexecute"
	"github.com/turbot/steampipe/workspace"
)

type Server struct {
	context   context.Context
	webSocket *melody.Melody
	workspace *workspace.Workspace
	client    *db.Client
}

type ExecutionPayload struct {
	Action string                   `json:"action"`
	Report *reportexecute.ReportRun `json:"report"`
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

	server := &Server{
		client:    dbClient,
		context:   ctx,
		webSocket: webSocket,
		workspace: loadedWorkspace,
	}

	server.workspace.RegisterReportEventHandler(server.HandleWorkspaceUpdate)

	return server, nil
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
	StartAPI(s.context, s.webSocket, s.workspace, s.client)
}

func (s *Server) Shutdown() {
	// Close the DB client
	if s.client != nil {
		s.client.Close()
	}

	if s.webSocket != nil {
		s.webSocket.Close()
	}

	// Close the workspace
	if s.workspace != nil {
		s.workspace.Close()
	}
}

func (s *Server) HandleWorkspaceUpdate(event reporteventpublisher.ReportEvent) {
	// TODO ...
	fmt.Println("Got update event", event)
	switch e := event.(type) {
	case *reportevents.ExecutionStarted:
		fmt.Println("Got execution started event", *e)
		s.webSocket.Broadcast(buildExecutionStartedPayload(e))
	case *reportevents.ExecutionComplete:
		fmt.Println("Got execution complete event", *e)
		s.webSocket.Broadcast(buildExecutionCompletePayload(e))
	}
}
