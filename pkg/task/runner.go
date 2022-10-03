package task

import (
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/db/db_local"
	"github.com/turbot/steampipe/pkg/error_helpers"
	"github.com/turbot/steampipe/pkg/statefile"
	"github.com/turbot/steampipe/pkg/utils"
)

const minimumDurationBetweenChecks = 24 * time.Hour

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
		runJobAsync(func() {
			versionNotificationLines = checkSteampipeVersion(r.currentState.InstallationID)
		}, &waitGroup)

		// check whether an updated version is available
		runJobAsync(func() {
			pluginNotificationLines = checkPluginVersions(r.currentState.InstallationID)
		}, &waitGroup)

		// remove log files older than 7 days
		runJobAsync(func() { db_local.TrimLogs() }, &waitGroup)

		// validate and regenerate service SSL certificates
		runJobAsync(func() { validateServiceCertificates() }, &waitGroup)

		// wait for all jobs to complete
		waitGroup.Wait()

		// display notifications, if any
		notificationLines := append(versionNotificationLines, pluginNotificationLines...)
		if len(notificationLines) > 0 {
			displayUpdateNotification(notificationLines)
		}

		// save the state - this updates the last checked time
		if err := r.currentState.Save(); err != nil {
			error_helpers.ShowWarning(fmt.Sprintf("Regular task runner failed to save state file: %s", err))
		}
	}
}

func runJobAsync(job func(), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		job()
		wg.Done()
	}()
}

// determines whether the task runner should run at all
// tasks are to be run at most once every 24 hours
// also, this is not to run in batch query mode
func (r *Runner) shouldRun() bool {
	utils.LogTime("task.Runner.shouldRun start")
	defer utils.LogTime("task.Runner.shouldRun end")

	cmd := viper.Get(constants.ConfigKeyActiveCommand).(*cobra.Command)
	cmdArgs := viper.GetStringSlice(constants.ConfigKeyActiveCommandArgs)
	if isIgnoredCmd(cmd, cmdArgs) {
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
	durationElapsedSinceLastCheck := now.Sub(lastCheckedAt)

	return durationElapsedSinceLastCheck > minimumDurationBetweenChecks
}

func isIgnoredCmd(cmd *cobra.Command, cmdArgs []string) bool {
	return (isPluginUpdateCmd(cmd) ||
		isPluginManagerCmd(cmd) ||
		isServiceStopCmd(cmd) ||
		isBatchQueryCmd(cmd, cmdArgs) ||
		isCompletionCmd(cmd))
}

func isServiceStopCmd(cmd *cobra.Command) bool {
	return cmd.Parent() != nil && cmd.Parent().Name() == "service" && cmd.Name() == "stop"
}
func isCompletionCmd(cmd *cobra.Command) bool {
	return cmd.Name() == "completion"
}
func isPluginManagerCmd(cmd *cobra.Command) bool {
	return cmd.Name() == "plugin-manager"
}
func isPluginUpdateCmd(cmd *cobra.Command) bool {
	return cmd.Name() == "update" && cmd.Parent() != nil && cmd.Parent().Name() == "plugin"
}
func isBatchQueryCmd(cmd *cobra.Command, cmdArgs []string) bool {
	return cmd.Name() == "query" && len(cmdArgs) > 0
}
