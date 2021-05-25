package reportserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/turbot/steampipe/executionlayer"

	"github.com/turbot/go-kit/types"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
	"gopkg.in/olahol/melody.v1"
)

type ClientRequestReportPayload struct {
	FullName string `json:"full_name"`
}

type ClientRequestPayload struct {
	Report ClientRequestReportPayload `json:"report"`
}

type ClientRequest struct {
	Action  string               `json:"action"`
	Payload ClientRequestPayload `json:"payload"`
}

type AvailableReportsPayload struct {
	Action  string            `json:"action"`
	Reports map[string]string `json:"reports"`
}

func availableReportsPayload(reports map[string]*modconfig.Report) []byte {
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

func Init(ctx context.Context, webSocket *melody.Melody, workspace *workspace.Workspace) {
	// Return list of reports on connect
	webSocket.HandleConnect(func(session *melody.Session) {
		fmt.Println("Client connected")
		//reports := listReportsForWorkspace()
		//err := webSocket.Broadcast(availableReportsPayload(reports))
		//if err != nil {
		//	log.Println(err)
		//}
	})

	webSocket.HandleMessage(func(s *melody.Session, msg []byte) {
		fmt.Println("Got message", string(msg))
		var request ClientRequest
		if err := json.Unmarshal(msg, &request); err != nil {
			// what???
			// TODO how to handle error
		} else {
			switch request.Action {
			case "available_reports":
				reports := workspace.Mod.Reports
				webSocket.Broadcast(availableReportsPayload(reports))
			case "select_report":
				fmt.Println(fmt.Sprintf("Got event: %v", request.Payload.Report))
				//for reportName := range workspace.ReportMap {
				executionlayer.ExecuteReport(ctx, request.Payload.Report.FullName, workspace)
				//break
				//}
				//report := workspace.Mod.Reports[request.Payload.Report.FullName]
				//go reportevents.GenerateReportEvents(report, executorFunction)
			}
		}
	})
}
