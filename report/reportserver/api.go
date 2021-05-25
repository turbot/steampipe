package reportserver

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/turbot/steampipe/report/reporteventpublisher"
	"github.com/turbot/steampipe/workspace"

	"gopkg.in/olahol/melody.v1"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func StartAPI(webSocket *melody.Melody, workspace *workspace.Workspace, handler reporteventpublisher.ReportEventHandler) {
	router := gin.Default()

	go Init(webSocket, workspace, handler)

	router.Use(static.Serve("/", static.LocalFile("./static", true)))

	router.GET("/ws", func(c *gin.Context) {
		webSocket.HandleRequest(c.Writer, c.Request)
	})

	srv := &http.Server{
		Addr:    ":5000",
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
