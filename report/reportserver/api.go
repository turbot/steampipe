package reportserver

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/turbot/steampipe/filepaths"
	"gopkg.in/olahol/melody.v1"
)

// https://stackoverflow.com/questions/39320371/how-start-web-server-to-open-page-in-browser-in-golang
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func StartAPI(ctx context.Context, webSocket *melody.Melody) {
	router := gin.Default()

	assetsDirectory := filepaths.ReportAssetsPath()

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

	_ = openBrowser("http://localhost:3001")

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
