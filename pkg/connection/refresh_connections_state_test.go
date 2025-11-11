package connection

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/turbot/pipe-fittings/v2/error_helpers"
	pfplugin "github.com/turbot/pipe-fittings/v2/plugin"
	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/pluginmanager_service/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/steampipeconfig"
)

// mockPluginManager is a minimal mock implementation for testing
type mockPluginManager struct {
	pool *pgxpool.Pool
}

func (m *mockPluginManager) Pool() *pgxpool.Pool {
	return m.pool
}

func (m *mockPluginManager) OnConnectionConfigChanged(context.Context, ConnectionConfigMap, map[string]*pfplugin.Plugin) {
}

func (m *mockPluginManager) GetConnectionConfig() ConnectionConfigMap {
	return nil
}

func (m *mockPluginManager) HandlePluginLimiterChanges(PluginLimiterMap) error {
	return nil
}

func (m *mockPluginManager) ShouldFetchRateLimiterDefs() bool {
	return false
}

func (m *mockPluginManager) LoadPluginRateLimiters(map[string]string) (PluginLimiterMap, error) {
	return nil, nil
}

func (m *mockPluginManager) SendPostgresSchemaNotification(context.Context) error {
	return nil
}

func (m *mockPluginManager) SendPostgresErrorsAndWarningsNotification(context.Context, error_helpers.ErrorAndWarnings) {
}

func (m *mockPluginManager) UpdatePluginColumnsTable(context.Context, map[string]*sdkproto.Schema, []string) error {
	return nil
}

func (m *mockPluginManager) Get(req *proto.GetRequest) (*proto.GetResponse, error) {
	return nil, nil
}

func (m *mockPluginManager) RefreshConnections(req *proto.RefreshConnectionsRequest) (*proto.RefreshConnectionsResponse, error) {
	return nil, nil
}

func (m *mockPluginManager) Shutdown(req *proto.ShutdownRequest) (*proto.ShutdownResponse, error) {
	return nil, nil
}

// TestRefreshConnectionState_ConnectionOrderEdgeCases tests edge cases in connection ordering
// This test demonstrates issue #4779 - GlobalConfig nil check in newRefreshConnectionState
// The code at line 75 calls GlobalConfig.GetNonSearchPathConnections without checking if GlobalConfig is nil
func TestRefreshConnectionState_ConnectionOrderEdgeCases(t *testing.T) {
	// Save original GlobalConfig
	originalGlobalConfig := steampipeconfig.GlobalConfig
	defer func() {
		// Restore original GlobalConfig
		steampipeconfig.GlobalConfig = originalGlobalConfig
	}()

	// Set GlobalConfig to nil to trigger the bug
	steampipeconfig.GlobalConfig = nil

	ctx := context.Background()

	// Create a mock plugin manager with a nil pool
	mockPM := &mockPluginManager{pool: nil}

	// After the fix, this should return an error instead of panicking
	// Before the fix, this will panic with nil pointer dereference at line 75
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Function panicked when it should return an error: %v", r)
		}
	}()

	result, err := newRefreshConnectionState(ctx, mockPM, nil)

	// We expect an error (not a panic) when GlobalConfig is nil
	if err == nil {
		t.Fatal("Expected an error when GlobalConfig is nil, but got nil")
	}

	if result != nil {
		t.Fatal("Expected nil result when GlobalConfig is nil, but got non-nil")
	}
}
