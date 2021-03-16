package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/go-hclog"
	_ "github.com/lib/pq"
	"github.com/turbot/steampipe-plugin-sdk/logging"
	"github.com/turbot/steampipe/cmd"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/task"
	"github.com/turbot/steampipe/utils"
)

var Logger hclog.Logger

func main() {

	checkRoot()

	/// setup logging
	logging.LogTime("start")
	createLogger()
	log.Println("[TRACE] tracing enabled")

	cmd.InitCmd()

	// run periodic tasks - update check and log clearing
	task.NewRunner().Run()

	// execute the command
	cmd.Execute()

	// remove the temp directory
	// don't care if it could not be removed
	defer os.RemoveAll(constants.TempDir())

	logging.LogTime("end")
	utils.DisplayProfileData()
}

// CreateLogger :: create a hclog logger with the level specified by the SP_LOG env var
func createLogger() {
	// TODO GET FROM VIPER
	level := logging.LogLevel()

	options := &hclog.LoggerOptions{Name: "steampipe", Level: hclog.LevelFromString(level)}
	if options.Output == nil {
		options.Output = os.Stderr
	}
	Logger = hclog.New(options)
	log.SetOutput(Logger.StandardWriter(&hclog.StandardLoggerOptions{InferLevels: true}))
	log.SetPrefix("")
	log.SetFlags(0)
}

// this is to replicate the user security mechanism of out underlying
// postgresql engine.
func checkRoot() {
	if os.Geteuid() == 0 {
		utils.ShowError(fmt.Errorf(`Steampipe cannot be run as the "root" user.
To reduce security risk, use an unprivileged user account instead.`))

		os.Exit(-1)
	}

	/*
	 * Also make sure that real and effective uids are the same. Executing as
	 * a setuid program from a root shell is a security hole, since on many
	 * platforms a nefarious subroutine could setuid back to root if real uid
	 * is root.  (Since nobody actually uses postgres as a setuid program,
	 * trying to actively fix this situation seems more trouble than it's
	 * worth; we'll just expend the effort to check for it.)
	 */

	if os.Geteuid() != os.Getuid() {
		utils.ShowError(fmt.Errorf("real and effective user IDs must match."))
		os.Exit(-1)
	}
}
