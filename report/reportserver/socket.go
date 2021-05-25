package reportserver

import (
	"encoding/json"
	"fmt"
	"github.com/turbot/steampipe/report/reportevents"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/workspace"
	"gopkg.in/olahol/melody.v1"
)

type ClientRequestPayload struct {
	Report *modconfig.Report `json:"report"`
}

type ClientRequest struct {
	Action  string               `json:"action"`
	Payload ClientRequestPayload `json:"payload"`
}

type AvailableReportsPayload struct {
	Action  string                       `json:"action"`
	Reports map[string]*modconfig.Report `json:"reports"`
}

func availableReportsPayload(reports map[string]*modconfig.Report) []byte {
	payload := AvailableReportsPayload{
		Action:  "available_reports",
		Reports: reports,
	}
	jsonString, _ := json.Marshal(payload)
	return jsonString
}

func Init(webSocket *melody.Melody, workspace *workspace.Workspace, executorFunction reportevents.ExecutorFunction) {
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
				reports := workspace.ReportMap
				webSocket.Broadcast(availableReportsPayload(reports))
			case "select_report":
				fmt.Println(fmt.Sprintf("Got event: %v", *request.Payload.Report))
				go reportevents.GenerateReportEvents(request.Payload.Report, executorFunction)
			}
		}
	})
}
