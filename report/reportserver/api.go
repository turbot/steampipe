package reportserver

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
)

func StartAPI(ctx context.Context, webSocket *melody.Melody) {
	router := gin.Default()

	_, filename, _, _ := runtime.Caller(1)
	assetsDirectory := path.Join(path.Dir(filename), "assets")

	router.Use(static.Serve("/", static.LocalFile(assetsDirectory, true)))

	router.GET("/ws", func(c *gin.Context) {
		webSocket.HandleRequest(c.Writer, c.Request)
	})

	router.NoRoute(func(c *gin.Context) {
		//c.File("./static/index.html")
		c.File(path.Join(assetsDirectory, "index.html"))
	})

	srv := &http.Server{
		Addr:    ":3001",
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

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
