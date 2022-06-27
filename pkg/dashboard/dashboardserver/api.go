package dashboardserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path"
	"runtime"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/filepaths"
	"github.com/turbot/steampipe/pkg/utils"
	"gopkg.in/olahol/melody.v1"
)

// https://stackoverflow.com/questions/39320371/how-start-web-server-to-open-page-in-browser-in-golang
func OpenBrowser(url string) error {
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

func startAPIAsync(ctx context.Context, webSocket *melody.Melody) chan struct{} {
	doneChan := make(chan struct{})

	go func() {
		gin.SetMode(gin.ReleaseMode)
		router := gin.New()
		// only add the Recovery middleware
		router.Use(gin.Recovery())

		assetsDirectory := filepaths.EnsureDashboardAssetsDir()

		router.Use(static.Serve("/", static.LocalFile(assetsDirectory, true)))

		router.GET("/ws", func(c *gin.Context) {
			webSocket.HandleRequest(c.Writer, c.Request)
		})

		router.NoRoute(func(c *gin.Context) {
			c.File(path.Join(assetsDirectory, "index.html"))
		})

		dashboardServerPort := viper.GetInt(constants.ArgDashboardPort)
		dashboardServerListen := "localhost"
		if viper.GetString(constants.ArgDashboardListen) == string(ListenTypeNetwork) {
			dashboardServerListen = ""
		}

		srv := &http.Server{
			Addr:    fmt.Sprintf("%s:%d", dashboardServerListen, dashboardServerPort),
			Handler: router,
		}

		go func() {
			// service connections
			if err := srv.ListenAndServe(); err != nil {
				log.Printf("listen: %s\n", err)
			}
		}()

		outputReady(ctx, fmt.Sprintf("Dashboard server started on %d and listening on %s", dashboardServerPort, viper.GetString(constants.ArgDashboardListen)))
		OutputMessage(ctx, fmt.Sprintf("Visit http://localhost:%d", dashboardServerPort))
		OutputMessage(ctx, "Press Ctrl+C to exit")
		<-ctx.Done()
		log.Println("Shutdown Server ...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := webSocket.Close(); err != nil {
			utils.ShowErrorWithMessage(ctx, err, "Websocket shutdown failed")
		}

		if err := srv.Shutdown(shutdownCtx); err != nil {
			utils.ShowErrorWithMessage(ctx, err, "Server shutdown failed")
		}
		log.Println("[TRACE] Server exiting")

		// indicate the API server is done
		doneChan <- struct{}{}
	}()

	return doneChan
}
