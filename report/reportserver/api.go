package reportserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path"
	"runtime"
	"time"

	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"

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

func StartAPI(ctx context.Context, webSocket *melody.Melody) *http.Server {
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

	reportServerPort := viper.GetInt(constants.ArgReportServerPort)
	reportServerListen := "localhost"
	if viper.GetString(constants.ArgReportServerListen) == string(ListenTypeNetwork) {
		reportServerListen = ""
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", reportServerListen, reportServerPort),
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	_ = openBrowser(fmt.Sprintf("http://localhost:%d", reportServerPort))
	<-ctx.Done()
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := webSocket.Close(); err != nil {
		fmt.Println("WebSocket Shutdown:", err)
	}

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Println("Server Shutdown:", err)
	}
	fmt.Println("Server exiting")

	return srv
}
