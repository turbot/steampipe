package reportserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/turbot/steampipe/db/db_common"
	"github.com/turbot/steampipe/report/reportexecute"
	"github.com/turbot/steampipe/workspace"
	"gopkg.in/olahol/melody.v1"
)

type ClientRequestDashboardPayload struct {
	FullName string `json:"full_name"`
}

type ClientRequestPayload struct {
	Dashboard ClientRequestDashboardPayload `json:"dashboard"`
}

type ClientRequest struct {
	Action  string               `json:"action"`
	Payload ClientRequestPayload `json:"payload"`
}

type ModAvailableDashboard struct {
	Title     string `json:"title,omitempty"`
	FullName  string `json:"full_name"`
	ShortName string `json:"short_name"`
}

type AvailableDashboardsPayload struct {
	Action          string                                      `json:"action"`
	DashboardsByMod map[string]map[string]ModAvailableDashboard `json:"dashboards_by_mod"`
}

type ModDashboardMetadata struct {
	Title     string `json:"title,omitempty"`
	FullName  string `json:"full_name"`
	ShortName string `json:"short_name"`
}

type DashboardMetadata struct {
	Mod           ModDashboardMetadata            `json:"mod"`
	InstalledMods map[string]ModDashboardMetadata `json:"installed_mods,omitempty"`
}

type DashboardMetadataPayload struct {
	Action   string            `json:"action"`
	Metadata DashboardMetadata `json:"metadata"`
}

func Init(ctx context.Context, webSocket *melody.Melody, workspace *workspace.Workspace, dbClient db_common.Client, socketSessions map[*melody.Session]*ReportClientInfo, mutex *sync.Mutex) {
	// Return list of reports on connect
	webSocket.HandleConnect(func(session *melody.Session) {
		log.Println("[TRACE] Client connected")
		mutex.Lock()
		socketSessions[session] = &ReportClientInfo{}
		mutex.Unlock()
	})

	webSocket.HandleDisconnect(func(session *melody.Session) {
		log.Println("[TRACE] Client disconnected")
		mutex.Lock()
		delete(socketSessions, session)
		mutex.Unlock()
	})

	webSocket.HandleMessage(func(session *melody.Session, msg []byte) {
		log.Println("[TRACE] Got message", string(msg))
		var request ClientRequest
		// if we could not decode message - ignore
		if err := json.Unmarshal(msg, &request); err == nil {

			switch request.Action {
			case "get_dashboard_metadata":
				payload, err := buildDashboardMetadataPayload(workspace)
				if err != nil {
					panic(fmt.Errorf("error building payload for get_metadata: %v", err))
				}
				session.Write(payload)
			case "get_available_dashboards":
				payload, err := buildAvailableDashboardsPayload(workspace)
				if err != nil {
					panic(fmt.Errorf("error building payload for get_available_dashboards: %v", err))
				}
				session.Write(payload)
			case "select_dashboard":
				log.Printf("[TRACE] Got event: %v\n", request.Payload.Dashboard)
				mutex.Lock()
				reportClientInfo := socketSessions[session]
				reportClientInfo.Report = &request.Payload.Dashboard.FullName
				mutex.Unlock()
				reportexecute.ExecuteReportNode(ctx, request.Payload.Dashboard.FullName, workspace, dbClient)
			}
		}
	})
}
