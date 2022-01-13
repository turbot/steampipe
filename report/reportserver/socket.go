package reportserver

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/report/reportexecute"
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

func Init(ctx context.Context, webSocket *melody.Melody, workspace *workspace.Workspace, dbClient db_common.Client, socketSessions map[*melody.Session]*ReportClientInfo, mutex *sync.Mutex) {
	// Return list of reports on connect
	webSocket.HandleConnect(func(session *melody.Session) {
		fmt.Println("Client connected")
		mutex.Lock()
		socketSessions[session] = &ReportClientInfo{}
		mutex.Unlock()
	})

	webSocket.HandleDisconnect(func(session *melody.Session) {
		fmt.Println("Client disconnected")
		mutex.Lock()
		delete(socketSessions, session)
		mutex.Unlock()
	})

	webSocket.HandleMessage(func(session *melody.Session, msg []byte) {
		fmt.Println("Got message", string(msg))
		var request ClientRequest
		if err := json.Unmarshal(msg, &request); err != nil {
			// what???
			// TODO how to handle error
		} else {
			switch request.Action {
			case "available_reports":
				reports := workspace.Mod.Reports
				session.Write(buildAvailableReportsPayload(reports))
			case "select_report":
				fmt.Printf("Got event: %v\n", request.Payload.Report)
				mutex.Lock()
				reportClientInfo := socketSessions[session]
				reportClientInfo.Report = &request.Payload.Report.FullName
				mutex.Unlock()
				reportexecute.ExecuteReportNode(ctx, request.Payload.Report.FullName, workspace, dbClient)
			}
		}
	})
}
