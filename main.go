package main

import (

	// need to attach this driver to the default sql package

	"github.com/turbot/steampipe/task"
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
	_ "github.com/lib/pq"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmd"
	"github.com/turbot/steampipe/utils"
)

var Logger hclog.Logger

func main() {
	logging.LogTime("start")
	createLogger()
	log.Println("[TRACE] tracing enabled")

	cmd.Execute()

	logging.LogTime("end")
	utils.DisplayProfileData()

}

// CreateLogger :: create a hclog logger with the level specified by the SP_LOG env var
func createLogger() {
	level, ok := os.LookupEnv("SP_LOG")
	if !ok {
		level = "WARNING"
	}

	options := &hclog.LoggerOptions{Name: "steampipe", Level: hclog.LevelFromString(level)}

	if options.Output == nil {
		options.Output = os.Stderr
	}

	Logger = hclog.New(options)

	log.SetOutput(Logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)

}

// run period tasks - update check and log clearing
func init() {
	task.NewRunner().Run()
}
