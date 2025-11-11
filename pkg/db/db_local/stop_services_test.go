package db_local

import (
	"context"
	"testing"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/v2/app_specific"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

// TestStopServicesIdempotent tests that StopServices can be called multiple times
// without errors, demonstrating idempotent behavior.
// The function should handle repeated calls gracefully, whether a service is
// running or not.
// See: https://github.com/turbot/steampipe/issues/4817
func TestStopServicesIdempotent(t *testing.T) {
	// Set up the install directory (required for file path operations)
	app_specific.InstallDir, _ = filehelpers.Tildefy("~/.steampipe")

	ctx := context.Background()

	// Call StopServices multiple times - should be idempotent
	// Each call should succeed even if the service is not running

	status1, err1 := StopServices(ctx, false, constants.InvokerQuery)
	if err1 != nil {
		t.Fatalf("First call to StopServices failed: %v", err1)
	}
	t.Logf("First call status: %d", status1)

	// Second call - should be idempotent
	status2, err2 := StopServices(ctx, false, constants.InvokerQuery)
	if err2 != nil {
		t.Errorf("Second call to StopServices failed: %v - function should be idempotent", err2)
	}
	t.Logf("Second call status: %d", status2)

	// Third call - further validation
	status3, err3 := StopServices(ctx, false, constants.InvokerQuery)
	if err3 != nil {
		t.Errorf("Third call to StopServices failed: %v - function should be idempotent", err3)
	}
	t.Logf("Third call status: %d", status3)
}
