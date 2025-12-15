package task

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/v2/plugin"
	"github.com/turbot/pipe-fittings/v2/utils"
	"github.com/turbot/steampipe/v2/pkg/db/db_local"
	"github.com/turbot/steampipe/v2/pkg/error_helpers"
	"github.com/turbot/steampipe/v2/pkg/filepaths"
	"github.com/turbot/steampipe/v2/pkg/installationstate"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

const minimumDurationBetweenChecks = 24 * time.Hour

type Runner struct {
	currentState installationstate.InstallationState
	options      *taskRunConfig
}

// RunTasks runs all tasks asynchronously
// returns a channel which is closed once all tasks are finished or the provided context is cancelled
func RunTasks(ctx context.Context, cmd *cobra.Command, args []string, options ...TaskRunOption) chan struct{} {
	utils.LogTime("task.RunTasks start")
	defer utils.LogTime("task.RunTasks end")

	config := newRunConfig()
	for _, o := range options {
		o(config)
	}

	doneChannel := make(chan struct{}, 1)
	runner := newRunner(config)

	// if there are any notifications from the previous run - display them
	if err := runner.displayNotifications(cmd, args); err != nil {
		log.Println("[TRACE] faced error displaying notifications:", err)
	}

	// asynchronously run the task runner
	go func(c context.Context) {
		defer close(doneChannel)
		// check if a legacy notifications file exists
		exists := files.FileExists(filepaths.LegacyNotificationsFilePath())
		if exists {
			log.Println("[TRACE] found legacy notification file. removing")
			// if the legacy file exists, remove it
			os.Remove(filepaths.LegacyNotificationsFilePath())
		}

		// if the legacy file existed, then we should enforce a run, since we need
		// to update the available version cache
		if runner.shouldRun() || exists {
			for _, hook := range config.preHooks {
				hook(c)
			}
			runner.run(c)
		}
	}(ctx)

	return doneChannel
}

func newRunner(config *taskRunConfig) *Runner {
	utils.LogTime("task.NewRunner start")
	defer utils.LogTime("task.NewRunner end")

	r := new(Runner)
	r.options = config

	state, err := installationstate.Load()
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

	var availableCliVersion *CLIVersionCheckResponse
	var availablePluginVersions map[string]plugin.PluginVersionCheckReport

	waitGroup := sync.WaitGroup{}

	if r.options.runUpdateCheck {
		// Only perform version checks if GlobalConfig is initialized
		// This can be nil during tests or unusual startup scenarios
		if steampipeconfig.GlobalConfig != nil {
			// check whether an updated version is available
			r.runJobAsync(ctx, func(c context.Context) {
				availableCliVersion, _ = fetchAvailableCLIVersion(ctx, r.currentState.InstallationID)
			}, &waitGroup)

			// check whether an updated version is available
			r.runJobAsync(ctx, func(ctx context.Context) {
				availablePluginVersions = plugin.GetAllUpdateReport(ctx, r.currentState.InstallationID, steampipeconfig.GlobalConfig.PluginVersions)
			}, &waitGroup)
		}
	}

	// remove log files older than 7 days
	r.runJobAsync(ctx, func(_ context.Context) { db_local.TrimLogs() }, &waitGroup)

	// wait for all jobs to complete
	waitGroup.Wait()

	// check if the context was cancelled before starting any FileIO
	if error_helpers.IsContextCanceled(ctx) {
		// if the context was cancelled, we don't want to do anything
		return
	}

	// save the notifications, if any
	if err := r.saveAvailableVersions(availableCliVersion, availablePluginVersions); err != nil {
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
		IsPluginManagerCmd(cmd) ||
		isServiceStopCmd(cmd) ||
		IsBatchQueryCmd(cmd, cmdArgs) ||
		isCompletionCmd(cmd) ||
		isPluginListCmd(cmd))
}

func isServiceStopCmd(cmd *cobra.Command) bool {
	return cmd.Parent() != nil && cmd.Parent().Name() == "service" && cmd.Name() == "stop"
}
func isCompletionCmd(cmd *cobra.Command) bool {
	return cmd.Name() == "completion"
}
func IsPluginManagerCmd(cmd *cobra.Command) bool {
	return cmd.Name() == "plugin-manager"
}
func isPluginUpdateCmd(cmd *cobra.Command) bool {
	return cmd.Name() == "update" && cmd.Parent() != nil && cmd.Parent().Name() == "plugin"
}
func IsBatchQueryCmd(cmd *cobra.Command, cmdArgs []string) bool {
	return cmd.Name() == "query" && len(cmdArgs) > 0
}
func isPluginListCmd(cmd *cobra.Command) bool {
	return cmd.Name() == "list" && cmd.Parent() != nil && cmd.Parent().Name() == "plugin"
}

func IsCheckCmd(cmd *cobra.Command) bool {
	return cmd.Name() == "check"
}

func IsDashboardCmd(cmd *cobra.Command) bool {
	return cmd.Name() == "dashboard"
}

func IsModCmd(cmd *cobra.Command) bool {
	parent := cmd.Parent()
	return parent.Name() == "mod"
}
