package reportserver

import (
	"fmt"
	"gopkg.in/olahol/melody.v1"
)

func Init(webSocket *melody.Melody) {
	// Return list of reports on connect
	webSocket.HandleConnect(func(session *melody.Session) {
		fmt.Println("Client connected")
		//reports := listReportsForWorkspace()
		//err := webSocket.Broadcast(availableReportsPayload(reports))
		//if err != nil {
		//	log.Println(err)
		//}
	})
}
