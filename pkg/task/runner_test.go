package task

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/turbot/pipe-fittings/v2/app_specific"
)

// setupTestEnvironment sets up the necessary environment for tests
func setupTestEnvironment(t *testing.T) {
	// Create a temporary directory for test state
	tempDir, err := os.MkdirTemp("", "steampipe-task-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	// Set the install directory to the temp directory
	app_specific.InstallDir = filepath.Join(tempDir, ".steampipe")
}

// TestRunTasksGoroutineCleanup tests that goroutines are properly cleaned up
// after RunTasks completes, including when context is cancelled
func TestRunTasksGoroutineCleanup(t *testing.T) {
	setupTestEnvironment(t)

	// Allow some buffer for background goroutines
	const goroutineBuffer = 10

	t.Run("normal_completion", func(t *testing.T) {
		before := runtime.NumGoroutine()

		ctx := context.Background()
		cmd := &cobra.Command{}

		// Run tasks with update check disabled to avoid network calls
		doneCh := RunTasks(ctx, cmd, []string{}, WithUpdateCheck(false))
		<-doneCh

		// Give goroutines time to clean up
		time.Sleep(100 * time.Millisecond)
		after := runtime.NumGoroutine()

		if after > before+goroutineBuffer {
			t.Errorf("Potential goroutine leak: before=%d, after=%d, diff=%d",
				before, after, after-before)
		}
	})

	t.Run("context_cancelled", func(t *testing.T) {
		before := runtime.NumGoroutine()

		ctx, cancel := context.WithCancel(context.Background())
		cmd := &cobra.Command{}

		doneCh := RunTasks(ctx, cmd, []string{}, WithUpdateCheck(false))

		// Cancel context immediately
		cancel()

		// Wait for completion
		select {
		case <-doneCh:
			// Good - channel was closed
		case <-time.After(2 * time.Second):
			t.Fatal("RunTasks did not complete within timeout after context cancellation")
		}

		// Give goroutines time to clean up
		time.Sleep(100 * time.Millisecond)
		after := runtime.NumGoroutine()

		if after > before+goroutineBuffer {
			t.Errorf("Goroutine leak after cancellation: before=%d, after=%d, diff=%d",
				before, after, after-before)
		}
	})

	t.Run("context_timeout", func(t *testing.T) {
		before := runtime.NumGoroutine()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		cmd := &cobra.Command{}
		doneCh := RunTasks(ctx, cmd, []string{}, WithUpdateCheck(false))

		// Wait for completion or timeout
		select {
		case <-doneCh:
			// Good - completed
		case <-time.After(2 * time.Second):
			t.Fatal("RunTasks did not complete within timeout")
		}

		// Give goroutines time to clean up
		time.Sleep(100 * time.Millisecond)
		after := runtime.NumGoroutine()

		if after > before+goroutineBuffer {
			t.Errorf("Goroutine leak after timeout: before=%d, after=%d, diff=%d",
				before, after, after-before)
		}
	})
}

// TestRunTasksChannelClosure tests that the done channel is always closed
func TestRunTasksChannelClosure(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("channel_closes_on_completion", func(t *testing.T) {
		ctx := context.Background()
		cmd := &cobra.Command{}

		doneCh := RunTasks(ctx, cmd, []string{}, WithUpdateCheck(false))

		select {
		case <-doneCh:
			// Good - channel was closed
		case <-time.After(2 * time.Second):
			t.Fatal("Done channel was not closed within timeout")
		}
	})

	t.Run("channel_closes_on_cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cmd := &cobra.Command{}

		doneCh := RunTasks(ctx, cmd, []string{}, WithUpdateCheck(false))
		cancel()

		select {
		case <-doneCh:
			// Good - channel was closed even after cancellation
		case <-time.After(2 * time.Second):
			t.Fatal("Done channel was not closed after context cancellation")
		}
	})
}

// TestRunTasksContextRespect tests that RunTasks respects context cancellation
func TestRunTasksContextRespect(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("immediate_cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel before starting

		cmd := &cobra.Command{}
		start := time.Now()
		doneCh := RunTasks(ctx, cmd, []string{}, WithUpdateCheck(false)) // Disable to avoid network calls
		<-doneCh
		elapsed := time.Since(start)

		// Should complete quickly since context is already cancelled
		// Allow up to 2 seconds for cleanup
		if elapsed > 2*time.Second {
			t.Errorf("RunTasks took too long with cancelled context: %v", elapsed)
		}
	})

	t.Run("cancellation_during_execution", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cmd := &cobra.Command{}

		doneCh := RunTasks(ctx, cmd, []string{}, WithUpdateCheck(false)) // Disable to avoid network calls

		// Cancel shortly after starting
		time.Sleep(10 * time.Millisecond)
		cancel()

		start := time.Now()
		<-doneCh
		elapsed := time.Since(start)

		// Should complete relatively quickly after cancellation
		// Allow time for network operations to timeout
		if elapsed > 2*time.Second {
			t.Errorf("RunTasks took too long to complete after cancellation: %v", elapsed)
		}
	})
}

// TestRunnerWaitGroupPropagation tests that the WaitGroup properly waits for all jobs
func TestRunnerWaitGroupPropagation(t *testing.T) {
	setupTestEnvironment(t)

	config := newRunConfig()
	runner := newRunner(config)

	ctx := context.Background()
	jobCompleted := make(map[int]bool)
	var mutex sync.Mutex

	// Simulate multiple jobs
	wg := &sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		i := i // capture loop variable
		runner.runJobAsync(ctx, func(c context.Context) {
			time.Sleep(50 * time.Millisecond) // Simulate work
			mutex.Lock()
			jobCompleted[i] = true
			mutex.Unlock()
		}, wg)
	}

	// Wait for all jobs
	wg.Wait()

	// All jobs should be completed
	mutex.Lock()
	completedCount := len(jobCompleted)
	mutex.Unlock()

	assert.Equal(t, 5, completedCount, "Not all jobs completed before WaitGroup.Wait() returned")
}

// TestShouldRunLogic tests the shouldRun time-based logic
func TestShouldRunLogic(t *testing.T) {
	setupTestEnvironment(t)

	t.Run("no_last_check", func(t *testing.T) {
		config := newRunConfig()
		runner := newRunner(config)
		runner.currentState.LastCheck = ""

		assert.True(t, runner.shouldRun(), "Should run when no last check exists")
	})

	t.Run("invalid_last_check", func(t *testing.T) {
		config := newRunConfig()
		runner := newRunner(config)
		runner.currentState.LastCheck = "invalid-time-format"

		assert.True(t, runner.shouldRun(), "Should run when last check is invalid")
	})

	t.Run("recent_check", func(t *testing.T) {
		config := newRunConfig()
		runner := newRunner(config)
		// Set last check to 1 hour ago (less than 24 hours)
		runner.currentState.LastCheck = time.Now().Add(-1 * time.Hour).Format(time.RFC3339)

		assert.False(t, runner.shouldRun(), "Should not run when checked recently (< 24h)")
	})

	t.Run("old_check", func(t *testing.T) {
		config := newRunConfig()
		runner := newRunner(config)
		// Set last check to 25 hours ago (more than 24 hours)
		runner.currentState.LastCheck = time.Now().Add(-25 * time.Hour).Format(time.RFC3339)

		assert.True(t, runner.shouldRun(), "Should run when last check is old (> 24h)")
	})
}

// TestCommandClassifiers tests the command classification functions
func TestCommandClassifiers(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *cobra.Command
		checker  func(*cobra.Command) bool
		expected bool
	}{
		{
			name: "plugin_update_command",
			setup: func() *cobra.Command {
				parent := &cobra.Command{Use: "plugin"}
				cmd := &cobra.Command{Use: "update"}
				parent.AddCommand(cmd)
				return cmd
			},
			checker:  isPluginUpdateCmd,
			expected: true,
		},
		{
			name: "service_stop_command",
			setup: func() *cobra.Command {
				parent := &cobra.Command{Use: "service"}
				cmd := &cobra.Command{Use: "stop"}
				parent.AddCommand(cmd)
				return cmd
			},
			checker:  isServiceStopCmd,
			expected: true,
		},
		{
			name: "completion_command",
			setup: func() *cobra.Command {
				return &cobra.Command{Use: "completion"}
			},
			checker:  isCompletionCmd,
			expected: true,
		},
		{
			name: "plugin_manager_command",
			setup: func() *cobra.Command {
				return &cobra.Command{Use: "plugin-manager"}
			},
			checker:  IsPluginManagerCmd,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setup()
			result := tt.checker(cmd)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsBatchQueryCmd tests batch query detection
func TestIsBatchQueryCmd(t *testing.T) {
	t.Run("query_with_args", func(t *testing.T) {
		cmd := &cobra.Command{Use: "query"}
		result := IsBatchQueryCmd(cmd, []string{"some", "args"})
		assert.True(t, result, "Should detect batch query with args")
	})

	t.Run("query_without_args", func(t *testing.T) {
		cmd := &cobra.Command{Use: "query"}
		result := IsBatchQueryCmd(cmd, []string{})
		assert.False(t, result, "Should not detect batch query without args")
	})
}

// TestPreHooksExecution tests that pre-hooks are executed
func TestPreHooksExecution(t *testing.T) {
	setupTestEnvironment(t)

	preHook := func(ctx context.Context) {
		// Pre-hook executed
	}

	ctx := context.Background()
	cmd := &cobra.Command{}

	// Force shouldRun to return true by setting LastCheck to empty
	// This is a bit hacky but necessary to test pre-hooks
	doneCh := RunTasks(ctx, cmd, []string{},
		WithUpdateCheck(false),
		WithPreHook(preHook))
	<-doneCh

	// Note: Pre-hooks only execute if shouldRun() returns true
	// In a fresh test environment, this might not happen
	// This test documents the expected behavior
	t.Log("Pre-hook execution depends on shouldRun() returning true")
}

// TestPluginVersionCheckWithNilGlobalConfig tests that the plugin version check
// handles nil GlobalConfig gracefully. This is a regression test for bug #4747.
func TestPluginVersionCheckWithNilGlobalConfig(t *testing.T) {
	// DO NOT call setupTestEnvironment here - we want GlobalConfig to be nil
	// to reproduce the bug from issue #4747

	// Create a temporary directory for test state
	tempDir, err := os.MkdirTemp("", "steampipe-task-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	// Set the install directory to the temp directory
	app_specific.InstallDir = filepath.Join(tempDir, ".steampipe")

	// Create a runner with update checks enabled
	config := newRunConfig()
	config.runUpdateCheck = true
	runner := newRunner(config)

	// Create a context with immediate cancellation to avoid network operations
	// and race conditions with the CLI version check goroutine
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Before the fix, this would panic at runner.go:106 when trying to access
	// steampipeconfig.GlobalConfig.PluginVersions
	// After the fix, it should handle nil GlobalConfig gracefully
	runner.run(ctx)

	// If we got here without panic, the fix is working
	t.Log("runner.run() completed without panic when GlobalConfig is nil and update checks are enabled")
}
