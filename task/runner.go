package task

import (
	"sync"
	"time"

	"github.com/turbot/steampipe/db/local_db"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/statefile"
	"github.com/turbot/steampipe/utils"
)

const minimumMinutesBetweenChecks = 1440 // 1 day

type Runner struct {
	currentState statefile.State
}

func RunTasks() {
	utils.LogTime("task.RunTasks start")
	defer utils.LogTime("task.RunTasks end")

	NewRunner().Run()
}

func NewRunner() *Runner {
	utils.LogTime("task.NewRunner start")
	defer utils.LogTime("task.NewRunner end")

	r := new(Runner)
	r.currentState, _ = statefile.LoadState()
	return r
}

func (r *Runner) Run() {
	utils.LogTime("task.Runner.Run start")
	defer utils.LogTime("task.Runner.Run end")

	var versionNotificationLines []string
	var pluginNotificationLines []string
	if r.shouldRun() {
		waitGroup := sync.WaitGroup{}

		// check whether an updated version is available
		waitGroup.Add(1)
		go r.runAsyncJob(func() {
			versionNotificationLines = checkSteampipeVersion(r.currentState.InstallationID)
		}, &waitGroup)

		// check whether an updated version is available
		waitGroup.Add(1)
		go r.runAsyncJob(func() {
			pluginNotificationLines = checkPluginVersions(r.currentState.InstallationID)
		}, &waitGroup)

		// remove log files older than 7 days
		waitGroup.Add(1)
		go r.runAsyncJob(func() { local_db.TrimLogs() }, &waitGroup)

		// wait for all jobs to complete
		waitGroup.Wait()

		// display notifications, if any
		notificationLines := append(versionNotificationLines, pluginNotificationLines...)
		if len(notificationLines) > 0 {
			displayUpdateNotification(notificationLines)
		}

		// save the state - this updates the last checked time
		r.currentState.Save()
	}
}

func (r *Runner) runAsyncJob(job func(), wg *sync.WaitGroup) {
	job()
	wg.Done()
}

// determines whether the task runner should run at all
// tasks are to be run at most once every 24 hours
// also, this is not to run in batch query mode
func (r *Runner) shouldRun() bool {
	utils.LogTime("task.Runner.shouldRun start")
	defer utils.LogTime("task.Runner.shouldRun end")

	cmd := viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)
	cmdArgs := viper.GetStringSlice(constants.ConfigKeyActiveCommandArgs)
	if cmd.Name() == "query" && len(cmdArgs) > 0 {
		// this is query batch mode
		// we will not run scheduled tasks in this mode
		return false
	}

	now := time.Now()
	if r.currentState.LastCheck == "" {
		return true
	}
	lastCheckedAt, err := time.Parse(time.RFC3339, r.currentState.LastCheck)
	if err != nil {
		return true
	}
	minutesElapsed := now.Sub(lastCheckedAt).Minutes()

	return minutesElapsed > minimumMinutesBetweenChecks
}
