package task

import (
	"context"
	"fmt"
	"log"
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

// RunTasks runs all tasks asynchronously
// returns a channel which is closed once all tasks are finished or the provided context is cancelled
func RunTasks(ctx context.Context, cmd *cobra.Command, args []string) chan struct{} {
	utils.LogTime("task.RunTasks start")
	defer utils.LogTime("task.RunTasks end")

	doneChannel := make(chan struct{}, 1)
	runner := newRunner()

	// if there are any notifications from the previous run - display them
	if err := runner.displayNotifications(cmd, args); err != nil {
		log.Println("[TRACE] faced error displaying notifications:", err)
	}

	// asynchronously run the task runner
	go func(c context.Context) {
		defer close(doneChannel)
		if runner.shouldRun() {
			runner.run(c)
		}
	}(ctx)

	return doneChannel
}

func newRunner() *Runner {
	utils.LogTime("task.NewRunner start")
	defer utils.LogTime("task.NewRunner end")

	r := new(Runner)

	state, err := statefile.LoadState()
	if err != nil {
		// this error should never happen
		// log this and carry on
		log.Println("[TRACE] error loading state,", err)
	}
	r.currentState = state
	return r
}

func (r *Runner) run(ctx context.Context) {
	utils.LogTime("task.Runner.Run start")
	defer utils.LogTime("task.Runner.Run end")

	var versionNotificationLines []string
	var pluginNotificationLines []string

	waitGroup := sync.WaitGroup{}

	if viper.GetBool(constants.ArgUpdateCheck) {
		// check whether an updated version is available
		r.runJobAsync(ctx, func(c context.Context) {
			versionNotificationLines = checkSteampipeVersion(c, r.currentState.InstallationID)
		}, &waitGroup)

		// check whether an updated version is available
		r.runJobAsync(ctx, func(c context.Context) {
			pluginNotificationLines = checkPluginVersions(c, r.currentState.InstallationID)
		}, &waitGroup)
	}

	// remove log files older than 7 days
	r.runJobAsync(ctx, func(context.Context) { db_local.TrimLogs() }, &waitGroup)

	// validate and regenerate service SSL certificates
	r.runJobAsync(ctx, func(context.Context) { validateServiceCertificates() }, &waitGroup)

	// wait for all jobs to complete
	waitGroup.Wait()

	// check if the context was cancelled before starting any FileIO
	if utils.IsContextCancelled(ctx) {
		// if the context was cancelled, we don't want to do anything
		return
	}

	// save the notifications, if any
	if err := r.saveNotifications(versionNotificationLines, pluginNotificationLines); err != nil {
		error_helpers.ShowWarning(fmt.Sprintf("Regular task runner failed to save pending notifications: %s", err))
	}

	// save the state - this updates the last checked time
	if err := r.currentState.Save(); err != nil {
		error_helpers.ShowWarning(fmt.Sprintf("Regular task runner failed to save state file: %s", err))
	}
}

func (r *Runner) runJobAsync(ctx context.Context, job func(context.Context), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		// do this as defer, so that it always fires - even if there's a panic
		defer wg.Done()
		job(ctx)
	}()
}

// determines whether the task runner should run at all
// tasks are to be run at most once every 24 hours
func (r *Runner) shouldRun() bool {
	utils.LogTime("task.Runner.shouldRun start")
	defer utils.LogTime("task.Runner.shouldRun end")

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

func showNotificationsForCommand(cmd *cobra.Command, cmdArgs []string) bool {
	return !(isPluginUpdateCmd(cmd) ||
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
