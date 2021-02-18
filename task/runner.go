package task

import (
	"sync"
	"time"

	"github.com/turbot/steampipe/db"
	"github.com/turbot/steampipe/statefile"
)

const minimumMinutesBetweenChecks = 1440 // 1 day

type Runner struct {
	currentState statefile.State
	shouldRun    bool
}

func NewRunner() *Runner {
	r := new(Runner)
	r.currentState, _ = statefile.LoadState()
	r.shouldRun = r.getShouldRun()
	return r
}

func (r *Runner) Run() {
	if r.shouldRun {
		waitGroup := sync.WaitGroup{}

		// check whether an updated version is available
		waitGroup.Add(1)
		go r.runAsyncJob(func() { checkSteampipeVersion(r.currentState.InstallationID) }, &waitGroup)

		// check whether an updated version is available
		waitGroup.Add(1)
		go r.runAsyncJob(func() { checkPluginVersions(r.currentState.InstallationID) }, &waitGroup)

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

func (r *Runner) getShouldRun() bool {
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
