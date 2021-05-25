package reportserver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/turbot/steampipe/report/reporteventpublisher"
	"github.com/turbot/steampipe/report/reportexecute"

	"github.com/turbot/steampipe/workspace"

	"gopkg.in/olahol/melody.v1"

	"github.com/turbot/steampipe/report/reportevents"
)

type Server struct {
	context   context.Context
	webSocket *melody.Melody
	workspace *workspace.Workspace
}

type ExecutionPayload struct {
	Action string                   `json:"action"`
	Report *reportexecute.ReportRun `json:"report"`
}

func NewServer(ctx context.Context, webSocket *melody.Melody, workspace *workspace.Workspace) Server {
	return Server{
		ctx,
		webSocket,
		workspace,
	}
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
	StartAPI(s.context, s.webSocket, s.workspace)
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
