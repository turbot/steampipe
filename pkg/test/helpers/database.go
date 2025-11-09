package helpers

import (
	"context"
	"testing"
	"time"

	"github.com/turbot/steampipe/v2/pkg/db/db_common"
	"github.com/turbot/steampipe/v2/pkg/test/mocks"
)

// CreateTestDatabaseClient creates a mock database client for testing
func CreateTestDatabaseClient(t *testing.T) *mocks.MockClient {
	t.Helper()

	// Create a mock client with sensible defaults
	client := &mocks.MockClient{}

	// Set up default behavior that most tests will need
	client.AcquireSessionFunc = func(ctx context.Context) *db_common.AcquireSessionResult {
		return &db_common.AcquireSessionResult{
			Session: &db_common.DatabaseSession{
				BackendPid: 12345,
				SearchPath: []string{"public"},
			},
		}
	}

	client.ServerSettingsFunc = func() *db_common.ServerSettings {
		return &db_common.ServerSettings{
			StartTime:        time.Now(),
			SteampipeVersion: "0.0.0-test",
			FdwVersion:       "0.0.0-test",
			CacheEnabled:     false,
			CacheMaxTtl:      300,
			CacheMaxSizeMb:   1024,
		}
	}

	client.GetRequiredSessionSearchPathFunc = func() []string {
		return []string{"public"}
	}

	return client
}

// CreateTestDatabaseSession creates a mock database session for testing
func CreateTestDatabaseSession(t *testing.T) *db_common.DatabaseSession {
	t.Helper()

	return &db_common.DatabaseSession{
		BackendPid: 12345,
		SearchPath: []string{"public"},
		Connection: nil, // Tests can set this if needed
	}
}

// CreateTestDatabasePath creates a temporary database directory for testing
// This is useful for tests that need to work with file-based database operations
func CreateTestDatabasePath(t *testing.T) string {
	t.Helper()

	// Use the filesystem helper to create a temp directory
	dir := CreateTempDir(t)

	// Create common database subdirectories
	CreateTestDir(t, dir, "db")
	CreateTestDir(t, dir, "logs")
	CreateTestDir(t, dir, "plugins")

	return dir
}
