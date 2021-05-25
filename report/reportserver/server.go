package reportserver

import (
	"encoding/json"
	"fmt"
	"github.com/turbot/steampipe/report/reportexecute"
	"github.com/turbot/steampipe/workspace"

	"gopkg.in/olahol/melody.v1"

	"github.com/turbot/steampipe/report/reportevents"
)

type Server struct {
	WebSocket *melody.Melody
	Workspace *workspace.Workspace
}

type ExecutionPayload struct {
	Action string                   `json:"action"`
	Report *reportexecute.ReportRun `json:"report"`
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
	StartAPI(s.WebSocket, s.Workspace, s.HandleWorkspaceUpdate)
}

func (s *Server) HandleWorkspaceUpdate(event reportevents.ReportEvent) {
	// TODO ...
	fmt.Println("Got update event", event)
	switch e := event.(type) {
	case *reportevents.ExecutionStarted:
		fmt.Println("Got execution started event", *e)
		s.WebSocket.Broadcast(buildExecutionStartedPayload(e))
	case *reportevents.ExecutionComplete:
		fmt.Println("Got execution complete event", *e)
		s.WebSocket.Broadcast(buildExecutionCompletePayload(e))
	}
}
