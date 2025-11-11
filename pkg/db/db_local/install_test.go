package db_local

import (
	"context"
	"sync"
	"testing"

	filehelpers "github.com/turbot/go-kit/files"
	"github.com/turbot/pipe-fittings/v2/app_specific"
)

func TestIsValidDatabaseName(t *testing.T) {
	tests := map[string]bool{
		"valid_name":  true,
		"_valid_name": true,
		"InvalidName": false,
		"123Invalid":  false,
	}

	for dbName, expectedResult := range tests {
		if actualResult := isValidDatabaseName(dbName); actualResult != expectedResult {
			t.Logf("Expected %t for %s, but for %t", expectedResult, dbName, actualResult)
			t.Fail()
		}
	}
}

func TestIsValidDatabaseName_EmptyString(t *testing.T) {
	// Test that isValidDatabaseName handles empty strings gracefully
	// An empty string should return false, not panic
	result := isValidDatabaseName("")
	if result != false {
		t.Errorf("Expected false for empty string, got %v", result)
	}
}

// TestEnsureDBInstalled_Concurrent tests concurrent calls to EnsureDBInstalled
// This test demonstrates the TOCTOU (Time-of-Check-Time-of-Use) race condition
// where checking if the DB is installed and then acting on that information are
// separate operations that can be interleaved by concurrent goroutines.
func TestEnsureDBInstalled_Concurrent(t *testing.T) {
	// Set up the install directory (required for file path operations)
	app_specific.InstallDir, _ = filehelpers.Tildefy("~/.steampipe")

	ctx := context.Background()

	// Launch multiple concurrent calls to EnsureDBInstalled
	// This simulates the race condition where multiple callers might check
	// IsDBInstalled() and then proceed with installation
	var wg sync.WaitGroup
	goroutines := 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// This pattern demonstrates the TOCTOU issue:
			// 1. Check if DB is installed (Time of Check)
			// 2. If not installed, ensure it's installed (Time of Use)
			// The race occurs between these two steps
			if !IsDBInstalled() {
				t.Logf("Goroutine %d: DB not installed, calling EnsureDBInstalled", id)
			}

			err := EnsureDBInstalled(ctx)
			if err != nil {
				t.Errorf("Goroutine %d: EnsureDBInstalled failed: %v", id, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify the database is installed after all concurrent calls
	if !IsDBInstalled() {
		t.Error("Expected database to be installed after concurrent EnsureDBInstalled calls")
	}
}
