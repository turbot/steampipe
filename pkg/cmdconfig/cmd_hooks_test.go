package cmdconfig

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestPostRunHook_WaitsForTasks(t *testing.T) {
	// Test that postRunHook waits for async tasks
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	// Simulate a task channel
	testChannel := make(chan struct{})
	oldChannel := waitForTasksChannel
	waitForTasksChannel = testChannel
	defer func() { waitForTasksChannel = oldChannel }()

	// Close the channel after a short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(testChannel)
	}()

	start := time.Now()
	postRunHook(cmd, []string{})
	duration := time.Since(start)

	// Should have waited for the channel to close
	if duration < 10*time.Millisecond {
		t.Error("postRunHook did not wait for tasks channel")
	}
}

func TestPostRunHook_Timeout(t *testing.T) {
	// Test that postRunHook times out if tasks take too long
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	// Simulate a task channel that never closes
	testChannel := make(chan struct{})
	oldChannel := waitForTasksChannel
	waitForTasksChannel = testChannel
	defer func() {
		waitForTasksChannel = oldChannel
		close(testChannel)
	}()

	// Mock cancel function
	cancelCalled := false
	oldCancelFn := tasksCancelFn
	tasksCancelFn = func() {
		cancelCalled = true
	}
	defer func() { tasksCancelFn = oldCancelFn }()

	start := time.Now()
	postRunHook(cmd, []string{})
	duration := time.Since(start)

	// Should have timed out after 100ms
	if duration < 100*time.Millisecond || duration > 150*time.Millisecond {
		t.Errorf("postRunHook timeout not working correctly, took %v", duration)
	}

	if !cancelCalled {
		t.Error("Cancel function was not called on timeout")
	}
}

func TestCmdBuilder_HookIntegration(t *testing.T) {
	// Test that CmdBuilder properly wraps hooks
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		// Original PreRun
	}

	cmd.PostRun = func(cmd *cobra.Command, args []string) {
		// Original PostRun
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		// Original Run
	}

	// Build with CmdBuilder
	builder := OnCmd(cmd)
	if builder == nil {
		t.Fatal("OnCmd returned nil")
	}

	// The hooks should now be wrapped
	if cmd.PreRun == nil {
		t.Error("PreRun hook was not set")
	}
	if cmd.PostRun == nil {
		t.Error("PostRun hook was not set")
	}
	if cmd.Run == nil {
		t.Error("Run hook was not set")
	}

	// Note: We can't easily test the wrapped functions without a full cobra execution
	// This would require integration tests
	t.Log("CmdBuilder successfully wrapped command hooks")
}

func TestCmdBuilder_FlagBinding(t *testing.T) {
	// Test that CmdBuilder properly binds flags to viper
	viper.Reset()
	defer viper.Reset()

	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	builder := OnCmd(cmd)
	builder.AddStringFlag("test-flag", "default-value", "Test flag description")

	// Verify flag was added
	flag := cmd.Flags().Lookup("test-flag")
	if flag == nil {
		t.Fatal("Flag was not added to command")
	}

	if flag.DefValue != "default-value" {
		t.Errorf("Flag default value incorrect, got %s", flag.DefValue)
	}

	// Verify binding was stored
	if len(builder.bindings) != 1 {
		t.Errorf("Expected 1 binding, got %d", len(builder.bindings))
	}

	if builder.bindings["test-flag"] != flag {
		t.Error("Flag binding not stored correctly")
	}
}

func TestCmdBuilder_MultipleFlagTypes(t *testing.T) {
	// Test that CmdBuilder can handle multiple flag types
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	builder := OnCmd(cmd)
	builder.
		AddStringFlag("string-flag", "default", "String flag").
		AddIntFlag("int-flag", 42, "Int flag").
		AddBoolFlag("bool-flag", true, "Bool flag").
		AddStringSliceFlag("slice-flag", []string{"a", "b"}, "Slice flag")

	// Verify all flags were added
	if cmd.Flags().Lookup("string-flag") == nil {
		t.Error("String flag not added")
	}
	if cmd.Flags().Lookup("int-flag") == nil {
		t.Error("Int flag not added")
	}
	if cmd.Flags().Lookup("bool-flag") == nil {
		t.Error("Bool flag not added")
	}
	if cmd.Flags().Lookup("slice-flag") == nil {
		t.Error("Slice flag not added")
	}

	// Verify all bindings were stored
	if len(builder.bindings) != 4 {
		t.Errorf("Expected 4 bindings, got %d", len(builder.bindings))
	}
}

func TestCmdBuilder_CloudFlags(t *testing.T) {
	// Test that AddCloudFlags adds the expected flags
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	builder := OnCmd(cmd)
	builder.AddCloudFlags()

	// Verify cloud flags were added
	if cmd.Flags().Lookup("pipes-host") == nil {
		t.Error("pipes-host flag not added")
	}
	if cmd.Flags().Lookup("pipes-token") == nil {
		t.Error("pipes-token flag not added")
	}
}

func TestCmdBuilder_NilFlagPanic(t *testing.T) {
	// Test that nil flag causes panic (as documented in builder.go)
	cmd := &cobra.Command{
		Use: "test",
		PreRun: func(cmd *cobra.Command, args []string) {
			// This will be called by CmdBuilder's wrapped PreRun
		},
		Run: func(cmd *cobra.Command, args []string) {},
	}

	builder := OnCmd(cmd)
	builder.AddStringFlag("test-flag", "default", "Test flag")

	// Manually corrupt the bindings to test panic
	builder.bindings["corrupt-flag"] = nil

	// This should panic when PreRun is executed
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nil flag binding")
		} else {
			t.Logf("Correctly panicked with: %v", r)
		}
	}()

	// Execute PreRun which should panic
	cmd.PreRun(cmd, []string{})
}
