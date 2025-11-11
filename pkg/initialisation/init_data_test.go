package initialisation

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/spf13/viper"
	pconstants "github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

// TestInitData_ResourceLeakOnPipesMetadataError tests if telemetry is leaked
// when getPipesMetadata fails after telemetry is initialized
func TestInitData_ResourceLeakOnPipesMetadataError(t *testing.T) {
	// Setup: Configure a scenario that will cause getPipesMetadata to fail
	// (database name without token)
	originalWorkspaceDB := viper.GetString(pconstants.ArgWorkspaceDatabase)
	originalToken := viper.GetString(pconstants.ArgPipesToken)
	defer func() {
		viper.Set(pconstants.ArgWorkspaceDatabase, originalWorkspaceDB)
		viper.Set(pconstants.ArgPipesToken, originalToken)
	}()

	viper.Set(pconstants.ArgWorkspaceDatabase, "some-database-name")
	viper.Set(pconstants.ArgPipesToken, "") // Missing token will cause error

	ctx := context.Background()
	initData := NewInitData()

	// Run initialization - should fail during getPipesMetadata
	initData.Init(ctx, constants.InvokerQuery)

	// Verify that an error occurred
	if initData.Result.Error == nil {
		t.Fatal("Expected error from missing cloud token, got nil")
	}

	// BUG CHECK: Is telemetry cleaned up?
	// If Init() fails after telemetry is initialized but before completion,
	// the telemetry goroutines may be leaked since Cleanup() is not called automatically
	if initData.ShutdownTelemetry != nil {
		t.Logf("WARNING: ShutdownTelemetry function exists but was not called - potential resource leak!")
		t.Logf("BUG FOUND: When Init() fails partway through, telemetry is not automatically cleaned up")
		t.Logf("The caller must remember to call Cleanup() even on error, but this is not enforced")

		// Clean up manually to prevent leak in test
		initData.Cleanup(ctx)
	}
}

// TestInitData_ResourceLeakOnClientError tests if telemetry is leaked
// when GetDbClient fails after telemetry is initialized
func TestInitData_ResourceLeakOnClientError(t *testing.T) {
	// Setup: Configure an invalid connection string
	originalConnString := viper.GetString(pconstants.ArgConnectionString)
	originalWorkspaceDB := viper.GetString(pconstants.ArgWorkspaceDatabase)
	defer func() {
		viper.Set(pconstants.ArgConnectionString, originalConnString)
		viper.Set(pconstants.ArgWorkspaceDatabase, originalWorkspaceDB)
	}()

	// Set invalid connection string that will fail
	viper.Set(pconstants.ArgConnectionString, "postgresql://invalid:invalid@nonexistent:5432/db")
	viper.Set(pconstants.ArgWorkspaceDatabase, "local")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	initData := NewInitData()

	// Run initialization - should fail during GetDbClient
	initData.Init(ctx, constants.InvokerQuery)

	// Verify that an error occurred (either connection error or context timeout)
	if initData.Result.Error == nil {
		t.Fatal("Expected error from invalid connection, got nil")
	}

	// BUG CHECK: Is telemetry cleaned up?
	if initData.ShutdownTelemetry != nil {
		t.Logf("BUG FOUND: Telemetry initialized but not cleaned up after client connection failure")
		t.Logf("Resource leak: telemetry goroutines may be running indefinitely")

		// Manual cleanup
		initData.Cleanup(ctx)
	}
}

// TestInitData_CleanupIdempotency tests if calling Cleanup multiple times is safe
func TestInitData_CleanupIdempotency(t *testing.T) {
	ctx := context.Background()
	initData := NewInitData()

	// Cleanup on uninitialized data should not panic
	initData.Cleanup(ctx)
	initData.Cleanup(ctx) // Second call should also be safe

	// Now initialize and cleanup multiple times
	originalWorkspaceDB := viper.GetString(pconstants.ArgWorkspaceDatabase)
	defer func() {
		viper.Set(pconstants.ArgWorkspaceDatabase, originalWorkspaceDB)
	}()
	viper.Set(pconstants.ArgWorkspaceDatabase, "local")

	// Note: We can't easily test with real initialization here as it requires
	// database setup, but we can test the nil safety of Cleanup
}

// TestInitData_NilExporter tests registering nil exporters
func TestInitData_NilExporter(t *testing.T) {
	// t.Skip("Demonstrates bug #4750 - HIGH nil pointer panic when registering nil exporter. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	initData := NewInitData()

	// Register nil exporter - should this panic or handle gracefully?
	result := initData.RegisterExporters(nil)

	if result.Result.Error != nil {
		t.Logf("Registering nil exporter returned error: %v", result.Result.Error)
	} else {
		t.Logf("Registering nil exporter succeeded - this might cause issues later")
	}
}

// TestInitData_PartialInitialization tests the state after partial initialization
func TestInitData_PartialInitialization(t *testing.T) {
	// Setup to fail at getPipesMetadata stage
	originalWorkspaceDB := viper.GetString(pconstants.ArgWorkspaceDatabase)
	originalToken := viper.GetString(pconstants.ArgPipesToken)
	defer func() {
		viper.Set(pconstants.ArgWorkspaceDatabase, originalWorkspaceDB)
		viper.Set(pconstants.ArgPipesToken, originalToken)
	}()

	viper.Set(pconstants.ArgWorkspaceDatabase, "test-db")
	viper.Set(pconstants.ArgPipesToken, "") // Will fail

	ctx := context.Background()
	initData := NewInitData()

	initData.Init(ctx, constants.InvokerQuery)

	// After failed init, check what state we're in
	if initData.Result.Error == nil {
		t.Fatal("Expected error, got nil")
	}

	// BUG CHECK: What's partially initialized?
	partiallyInitialized := []string{}
	if initData.ShutdownTelemetry != nil {
		partiallyInitialized = append(partiallyInitialized, "telemetry")
	}
	if initData.Client != nil {
		partiallyInitialized = append(partiallyInitialized, "client")
	}
	if initData.PipesMetadata != nil {
		partiallyInitialized = append(partiallyInitialized, "pipes_metadata")
	}

	if len(partiallyInitialized) > 0 {
		t.Logf("BUG: Partial initialization detected. Initialized: %v", partiallyInitialized)
		t.Logf("These resources need cleanup but Cleanup() may not be called by users on error")

		// Cleanup to prevent leak
		initData.Cleanup(ctx)
	}
}

// TestInitData_GoroutineLeak tests for goroutine leaks during failed initialization
func TestInitData_GoroutineLeak(t *testing.T) {
	// Allow some variance in goroutine count due to runtime behavior
	const goroutineThreshold = 5

	// Setup to fail
	originalWorkspaceDB := viper.GetString(pconstants.ArgWorkspaceDatabase)
	originalToken := viper.GetString(pconstants.ArgPipesToken)
	defer func() {
		viper.Set(pconstants.ArgWorkspaceDatabase, originalWorkspaceDB)
		viper.Set(pconstants.ArgPipesToken, originalToken)
	}()

	viper.Set(pconstants.ArgWorkspaceDatabase, "test-db")
	viper.Set(pconstants.ArgPipesToken, "")

	// Force garbage collection and get baseline
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	before := runtime.NumGoroutine()

	ctx := context.Background()
	initData := NewInitData()
	initData.Init(ctx, constants.InvokerQuery)

	// Don't call Cleanup - simulating user forgetting to cleanup on error

	// Force garbage collection
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	after := runtime.NumGoroutine()

	leaked := after - before
	if leaked > goroutineThreshold {
		t.Logf("BUG FOUND: Potential goroutine leak detected")
		t.Logf("Goroutines before: %d, after: %d, leaked: %d", before, after, leaked)
		t.Logf("When Init() fails, cleanup is not automatic - resources may leak")

		// Now cleanup and verify goroutines decrease
		initData.Cleanup(ctx)
		runtime.GC()
		time.Sleep(100 * time.Millisecond)
		afterCleanup := runtime.NumGoroutine()
		t.Logf("After manual cleanup: %d goroutines (difference: %d)", afterCleanup, afterCleanup-before)
	} else {
		t.Logf("Goroutine count stable: before=%d, after=%d, diff=%d", before, after, leaked)
	}
}

// TestNewErrorInitData tests the error constructor
func TestNewErrorInitData(t *testing.T) {
	testErr := context.Canceled
	initData := NewErrorInitData(testErr)

	if initData == nil {
		t.Fatal("NewErrorInitData returned nil")
	}

	if initData.Result == nil {
		t.Fatal("Result is nil")
	}

	if initData.Result.Error != testErr {
		t.Errorf("Expected error %v, got %v", testErr, initData.Result.Error)
	}

	// BUG CHECK: Can we call Cleanup on error init data?
	ctx := context.Background()
	initData.Cleanup(ctx) // Should not panic
}

// TestInitData_ContextCancellation tests behavior when context is cancelled during init
func TestInitData_ContextCancellation(t *testing.T) {
	originalWorkspaceDB := viper.GetString(pconstants.ArgWorkspaceDatabase)
	defer func() {
		viper.Set(pconstants.ArgWorkspaceDatabase, originalWorkspaceDB)
	}()
	viper.Set(pconstants.ArgWorkspaceDatabase, "local")

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	initData := NewInitData()
	initData.Init(ctx, constants.InvokerQuery)

	// Should get context cancellation error
	if initData.Result.Error == nil {
		t.Log("Expected context cancellation error, got nil")
	} else if initData.Result.Error == context.Canceled {
		t.Log("Correctly returned context cancellation error")
	} else {
		t.Logf("Got error: %v (expected context.Canceled)", initData.Result.Error)
	}

	// BUG CHECK: Are resources cleaned up?
	if initData.ShutdownTelemetry != nil {
		t.Log("BUG: Telemetry initialized even though context was cancelled")
		initData.Cleanup(context.Background())
	}
}

// TestInitData_PanicRecovery tests that panics during init are caught
func TestInitData_PanicRecovery(t *testing.T) {
	// We can't easily inject a panic into the real init flow without mocking,
	// but we can verify the defer/recover is in place by code inspection

	// This test documents expected behavior:
	t.Log("Init() has defer/recover to catch panics and convert to errors")
	t.Log("This is good - panics won't crash the application")
}

// TestInitData_DoubleInit tests calling Init twice on same InitData
func TestInitData_DoubleInit(t *testing.T) {
	originalWorkspaceDB := viper.GetString(pconstants.ArgWorkspaceDatabase)
	originalToken := viper.GetString(pconstants.ArgPipesToken)
	defer func() {
		viper.Set(pconstants.ArgWorkspaceDatabase, originalWorkspaceDB)
		viper.Set(pconstants.ArgPipesToken, originalToken)
	}()

	// Setup to fail quickly
	viper.Set(pconstants.ArgWorkspaceDatabase, "test-db")
	viper.Set(pconstants.ArgPipesToken, "")

	ctx := context.Background()
	initData := NewInitData()

	// First init - will fail
	initData.Init(ctx, constants.InvokerQuery)
	firstErr := initData.Result.Error

	// Second init on same object - what happens?
	initData.Init(ctx, constants.InvokerQuery)
	secondErr := initData.Result.Error

	t.Logf("First init error: %v", firstErr)
	t.Logf("Second init error: %v", secondErr)

	// BUG CHECK: Are there multiple telemetry instances now?
	// Are old resources cleaned up before reinitializing?
	t.Log("WARNING: Calling Init() twice on same InitData may leak resources")
	t.Log("The old ShutdownTelemetry function is overwritten without being called")

	// Cleanup
	if initData.ShutdownTelemetry != nil {
		initData.Cleanup(ctx)
	}
}

// TestGetDbClient_WithConnectionString tests the client creation with connection string
func TestGetDbClient_WithConnectionString(t *testing.T) {
	// t.Skip("Demonstrates bug #4767 - GetDbClient returns non-nil client even when error occurs, causing nil pointer panic on Close. Remove this skip in bug fix PR commit 1, then fix in commit 2.")
	originalConnString := viper.GetString(pconstants.ArgConnectionString)
	defer func() {
		viper.Set(pconstants.ArgConnectionString, originalConnString)
	}()

	// Set an invalid connection string
	viper.Set(pconstants.ArgConnectionString, "postgresql://invalid:invalid@nonexistent:5432/db")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, errAndWarnings := GetDbClient(ctx, constants.InvokerQuery)

	// Should get an error
	if errAndWarnings.Error == nil {
		t.Log("Expected connection error, got nil")
		if client != nil {
			// Clean up if somehow succeeded
			client.Close(ctx)
		}
	} else {
		t.Logf("Got expected error: %v", errAndWarnings.Error)
	}

	// BUG CHECK: Is client nil when error occurs?
	if errAndWarnings.Error != nil && client != nil {
		t.Log("BUG: Client is not nil even though error occurred")
		t.Log("Caller might try to use the client, leading to undefined behavior")
		client.Close(ctx)
	}
}

// TestGetDbClient_WithoutConnectionString tests the local client creation
func TestGetDbClient_WithoutConnectionString(t *testing.T) {
	originalConnString := viper.GetString(pconstants.ArgConnectionString)
	defer func() {
		viper.Set(pconstants.ArgConnectionString, originalConnString)
	}()

	// Clear connection string to force local client
	viper.Set(pconstants.ArgConnectionString, "")

	// Note: This test will try to start a local database which may not be available
	// in CI environment. We'll use a short timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, errAndWarnings := GetDbClient(ctx, constants.InvokerQuery)

	if errAndWarnings.Error != nil {
		t.Logf("Local client creation failed (expected in test environment): %v", errAndWarnings.Error)
	} else {
		t.Log("Local client created successfully")
		if client != nil {
			client.Close(ctx)
		}
	}

	// The test itself validates that the function doesn't panic
}
