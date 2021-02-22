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

	// make sure that we are not running as root
	checkRoot()

	/// setup logging
	logging.LogTime("start")
	createLogger()
	log.Println("[TRACE] tracing enabled")

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

func checkRoot() {
	// return (os.Getuid() == 0)
	if os.Geteuid() == 0 {
		utils.ShowError(fmt.Errorf(`"root" execution of %s is not permitted. 
%s must be started under an unprivileged user ID to prevent possible system security compromise`,
			constants.Bold("steampipe"),
			constants.Bold("steampipe"),
		))

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
		utils.ShowError(fmt.Errorf(`%s: real and effective user IDs must match`, constants.Bold("steampipe")))

		os.Exit(-1)
	}
}
