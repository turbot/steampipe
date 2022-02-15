package dashboardserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/turbot/steampipe/dashboard/dashboardexecute"
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

func (s *Server) Init(ctx context.Context) {
	// Return list of dashboards on connect
	s.webSocket.HandleConnect(func(session *melody.Session) {
		log.Println("[TRACE] Client connected")
		s.addSession(session)
	})

	s.webSocket.HandleDisconnect(func(session *melody.Session) {
		log.Println("[TRACE] Client disconnected")
		s.clearSession(session)
	})

	s.webSocket.HandleMessage(func(session *melody.Session, msg []byte) {
		log.Println("[TRACE] Got message", string(msg))
		var request ClientRequest
		// if we could not decode message - ignore
		if err := json.Unmarshal(msg, &request); err == nil {

			switch request.Action {
			case "get_dashboard_metadata":
				payload, err := buildDashboardMetadataPayload(s.workspaceResources)
				if err != nil {
					panic(fmt.Errorf("error building payload for get_metadata: %v", err))
				}
				session.Write(payload)
			case "get_available_dashboards":
				payload, err := buildAvailableDashboardsPayload(s.workspaceResources)
				if err != nil {
					panic(fmt.Errorf("error building payload for get_available_dashboards: %v", err))
				}
				session.Write(payload)
			case "select_dashboard":
				log.Printf("[TRACE] Got event: %v\n", request.Payload.Dashboard)
				dashboardClientInfo := s.getSession(session)
				dashboardClientInfo.Dashboard = &request.Payload.Dashboard.FullName
				dashboardexecute.ExecuteDashboardNode(ctx, request.Payload.Dashboard.FullName, s.workspace, s.dbClient)
			}
		}
	})
	outputMessage(ctx, "Initialization complete")
}

func (s *Server) getSession(session *melody.Session) *DashboardClientInfo {
	s.mutex.Lock()
	dashboardClientInfo := s.dashboardClients[session]
	s.mutex.Unlock()
	return dashboardClientInfo
}

func (s *Server) clearSession(session *melody.Session) {
	s.mutex.Lock()
	delete(s.dashboardClients, session)
	s.mutex.Unlock()
}

func (s *Server) addSession(session *melody.Session) {
	s.mutex.Lock()
	s.dashboardClients[session] = &DashboardClientInfo{}
	s.mutex.Unlock()
}
