package reportserver

import (
	"encoding/json"
	"fmt"

	"gopkg.in/olahol/melody.v1"

	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type Server struct {
	WebSocket *melody.Melody
}

type ExecutionStartedPayload struct {
	Action string
	Report *modconfig.Report
}

func buildExecutionStartedPayload(event *reportevents.ExecutionStarted) []byte {
	payload := ExecutionStartedPayload{
		Action: "execution_started",
		Report: event.Report.Report,
	}
	jsonString, _ := json.Marshal(payload)
	return jsonString
}

// Starts the API server
func (s *Server) Start() {
	StartAPI(s.WebSocket)
}

func (s *Server) HandleWorkspaceUpdate(event reportevents.ReportEvent) {
	// TODO ...
	switch e := event.(type) {
	case *reportevents.ExecutionStarted:
		fmt.Println("Got execution started event", *e)
		s.WebSocket.Broadcast(buildExecutionStartedPayload(e))
	}
}
