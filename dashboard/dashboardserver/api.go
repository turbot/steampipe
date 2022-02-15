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
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/utils"
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

func StartAPI(ctx context.Context, webSocket *melody.Melody) *http.Server {
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

	dashboardServerPort := viper.GetInt(constants.ArgDashboardServerPort)
	dashboardServerListen := "localhost"
	if viper.GetString(constants.ArgDashboardServerListen) == string(ListenTypeNetwork) {
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

	_ = openBrowser(fmt.Sprintf("http://localhost:%d", dashboardServerPort))
	outputReady(ctx, fmt.Sprintf("Dashboard server started on %d and listening on %s", dashboardServerPort, viper.GetString(constants.ArgDashboardServerListen)))
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

	return srv
}
