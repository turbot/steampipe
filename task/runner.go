package task

import (
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/constants"
	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/statefile"
)

const minimumMinutesBetweenChecks = 1440 // 1 day

type Runner struct {
	currentState statefile.State
}

func RunTasks() {
	NewRunner().Run()
}

func NewRunner() *Runner {
	r := new(Runner)
	r.currentState, _ = statefile.LoadState()
	return r
}

func (r *Runner) Run() {
	if r.shouldRun() {
		waitGroup := sync.WaitGroup{}

		if r.shouldRunUpdateChecks() {
			// check whether an updated version is available
			waitGroup.Add(1)
			go r.runAsyncJob(func() { checkSteampipeVersion(r.currentState.InstallationID) }, &waitGroup)

			// check whether an updated version is available
			waitGroup.Add(1)
			go r.runAsyncJob(func() { checkPluginVersions(r.currentState.InstallationID) }, &waitGroup)
		}

		// remove log files older than 7 days
		waitGroup.Add(1)
		go r.runAsyncJob(func() { db.TrimLogs() }, &waitGroup)

		// update last check time
		waitGroup.Wait()
		r.currentState.Save()
		r.currentState, _ = statefile.LoadState()
	}
}

func (r *Runner) runAsyncJob(job func(), wg *sync.WaitGroup) {
	job()
	wg.Done()
}

// determines whether the task runner should run at all
// tasks are to be run at most once every 24 hours
func (r *Runner) shouldRun() bool {
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

// returns whether to run update checks for the CLI
// and its installed plugins
// update-checks are to be disabled for batch query mode
func (r *Runner) shouldRunUpdateChecks() bool {
	cmd := viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)
	cmdArgs := viper.GetStringSlice(constants.ConfigKeyActiveCommandArgs)
	if cmd.Name() == "query" && len(cmdArgs) > 0 {
		// this is query batch mode
		// we will not run update checks in this mode
		return false
	}
	return true
}
