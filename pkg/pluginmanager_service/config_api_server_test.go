package pluginmanager_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/connection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeConfigRequest(t *testing.T, connection, plugin, shortName, config, instance string) *bytes.Buffer {
	t.Helper()
	body := SetConnectionConfigRequest{
		Connection:      connection,
		Plugin:          plugin,
		PluginShortName: shortName,
		Config:          config,
		PluginInstance:  instance,
	}
	b, err := json.Marshal(body)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

// Test 1: Add a new connection via the config API
func TestConfigAPIServer_SetConnectionConfig_NewConnection(t *testing.T) {
	pm := newTestPluginManager(t)

	req := httptest.NewRequest(http.MethodPost, "/v1/connection/config",
		makeConfigRequest(t, "aws_tenant_1", "hub.steampipe.io/plugins/turbot/aws@latest", "aws",
			`connection "aws_tenant_1" { regions = ["us-east-1"] }`, "aws"))
	w := httptest.NewRecorder()

	pm.handleSetConnectionConfig(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body SetConnectionConfigResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.True(t, body.Success)
	assert.Empty(t, body.Error)

	// Verify the connection was added to the config map
	pm.mut.RLock()
	cfg, ok := pm.connectionConfigMap["aws_tenant_1"]
	pm.mut.RUnlock()

	assert.True(t, ok, "connection should exist in connectionConfigMap")
	assert.Equal(t, "aws_tenant_1", cfg.Connection)
	assert.Equal(t, "aws", cfg.PluginShortName)
	assert.Equal(t, "aws", cfg.PluginInstance)
}

// Test 2: Update an existing connection's config
func TestConfigAPIServer_SetConnectionConfig_UpdateExisting(t *testing.T) {
	pm := newTestPluginManager(t)

	// Seed an existing connection
	pm.connectionConfigMap["aws_tenant_1"] = &sdkproto.ConnectionConfig{
		Connection:      "aws_tenant_1",
		Plugin:          "hub.steampipe.io/plugins/turbot/aws@latest",
		PluginShortName: "aws",
		Config:          `connection "aws_tenant_1" { regions = ["us-east-1"] }`,
		PluginInstance:  "aws",
	}

	// Update with new config
	newConfig := `connection "aws_tenant_1" { regions = ["us-west-2", "eu-west-1"] }`
	req := httptest.NewRequest(http.MethodPost, "/v1/connection/config",
		makeConfigRequest(t, "aws_tenant_1", "hub.steampipe.io/plugins/turbot/aws@latest", "aws", newConfig, "aws"))
	w := httptest.NewRecorder()

	pm.handleSetConnectionConfig(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body SetConnectionConfigResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.True(t, body.Success)

	// Verify the config was updated
	pm.mut.RLock()
	cfg := pm.connectionConfigMap["aws_tenant_1"]
	pm.mut.RUnlock()

	assert.Equal(t, newConfig, cfg.Config)
}

// Test 3: Missing required fields return 400
func TestConfigAPIServer_SetConnectionConfig_MissingFields(t *testing.T) {
	pm := newTestPluginManager(t)

	tests := []struct {
		name       string
		body       SetConnectionConfigRequest
		errContain string
	}{
		{
			name:       "missing connection",
			body:       SetConnectionConfigRequest{PluginShortName: "aws", Config: "x", PluginInstance: "aws"},
			errContain: "'connection' is required",
		},
		{
			name:       "missing plugin_short_name",
			body:       SetConnectionConfigRequest{Connection: "c1", Config: "x", PluginInstance: "aws"},
			errContain: "'plugin_short_name' is required",
		},
		{
			name:       "missing config",
			body:       SetConnectionConfigRequest{Connection: "c1", PluginShortName: "aws", PluginInstance: "aws"},
			errContain: "'config' is required",
		},
		{
			name:       "missing plugin_instance",
			body:       SetConnectionConfigRequest{Connection: "c1", PluginShortName: "aws", Config: "x"},
			errContain: "'plugin_instance' is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/v1/connection/config", bytes.NewBuffer(b))
			w := httptest.NewRecorder()

			pm.handleSetConnectionConfig(w, req)

			resp := w.Result()
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

			var body SetConnectionConfigResponse
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
			assert.False(t, body.Success)
			assert.Contains(t, body.Error, tt.errContain)
		})
	}
}

// Test 4: Invalid JSON returns 400
func TestConfigAPIServer_SetConnectionConfig_InvalidJSON(t *testing.T) {
	pm := newTestPluginManager(t)

	req := httptest.NewRequest(http.MethodPost, "/v1/connection/config",
		bytes.NewBufferString(`{invalid json`))
	w := httptest.NewRecorder()

	pm.handleSetConnectionConfig(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body SetConnectionConfigResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.False(t, body.Success)
	assert.Contains(t, body.Error, "invalid JSON")
}

// Test 5: GET request returns 405
func TestConfigAPIServer_SetConnectionConfig_MethodNotAllowed(t *testing.T) {
	pm := newTestPluginManager(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/connection/config", nil)
	w := httptest.NewRecorder()

	pm.handleSetConnectionConfig(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Result().StatusCode)
}

// Test 6: Concurrent updates don't race
func TestConfigAPIServer_SetConnectionConfig_Concurrent(t *testing.T) {
	pm := newTestPluginManager(t)

	var wg sync.WaitGroup
	numGoroutines := 20
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			connName := fmt.Sprintf("conn_%d", idx)
			body := makeConfigRequest(t, connName, "hub.steampipe.io/plugins/turbot/aws@latest", "aws",
				fmt.Sprintf(`connection "%s" { regions = ["us-east-1"] }`, connName), "aws")

			req := httptest.NewRequest(http.MethodPost, "/v1/connection/config", body)
			w := httptest.NewRecorder()

			pm.handleSetConnectionConfig(w, req)

			if w.Result().StatusCode != http.StatusOK {
				errors <- fmt.Errorf("connection %s: expected 200, got %d", connName, w.Result().StatusCode)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}

	// All connections should be in the map
	pm.mut.RLock()
	defer pm.mut.RUnlock()
	assert.Equal(t, numGoroutines, len(pm.connectionConfigMap))
}

// Test 7: Updating one connection preserves existing connections
func TestConfigAPIServer_SetConnectionConfig_PreservesExisting(t *testing.T) {
	pm := newTestPluginManager(t)

	// Seed two existing connections
	pm.connectionConfigMap = connection.ConnectionConfigMap{
		"conn_a": &sdkproto.ConnectionConfig{
			Connection:      "conn_a",
			Plugin:          "hub.steampipe.io/plugins/turbot/aws@latest",
			PluginShortName: "aws",
			Config:          `connection "conn_a" { regions = ["us-east-1"] }`,
			PluginInstance:  "aws",
		},
		"conn_b": &sdkproto.ConnectionConfig{
			Connection:      "conn_b",
			Plugin:          "hub.steampipe.io/plugins/turbot/k8s@latest",
			PluginShortName: "kubernetes",
			Config:          `connection "conn_b" { context = "minikube" }`,
			PluginInstance:  "kubernetes",
		},
	}

	// Update only conn_a
	req := httptest.NewRequest(http.MethodPost, "/v1/connection/config",
		makeConfigRequest(t, "conn_a", "hub.steampipe.io/plugins/turbot/aws@latest", "aws",
			`connection "conn_a" { regions = ["eu-west-1"] }`, "aws"))
	w := httptest.NewRecorder()

	pm.handleSetConnectionConfig(w, req)

	assert.Equal(t, http.StatusOK, w.Result().StatusCode)

	// Verify conn_a was updated
	pm.mut.RLock()
	defer pm.mut.RUnlock()

	assert.Equal(t, `connection "conn_a" { regions = ["eu-west-1"] }`, pm.connectionConfigMap["conn_a"].Config)

	// Verify conn_b is untouched
	assert.Equal(t, `connection "conn_b" { context = "minikube" }`, pm.connectionConfigMap["conn_b"].Config)
	assert.Equal(t, "kubernetes", pm.connectionConfigMap["conn_b"].PluginShortName)
}

// Test 8: buildMergedConfigMap creates a correct merged copy
func TestBuildMergedConfigMap_NewEntry(t *testing.T) {
	pm := newTestPluginManager(t)

	pm.connectionConfigMap = connection.ConnectionConfigMap{
		"existing": &sdkproto.ConnectionConfig{
			Connection:     "existing",
			PluginInstance: "aws",
		},
	}

	newCfg := &sdkproto.ConnectionConfig{
		Connection:     "new_conn",
		PluginInstance: "gcp",
	}

	merged := pm.buildMergedConfigMap(newCfg)

	assert.Len(t, merged, 2)
	assert.Equal(t, "existing", merged["existing"].Connection)
	assert.Equal(t, "new_conn", merged["new_conn"].Connection)

	// Original map should be unmodified
	assert.Len(t, pm.connectionConfigMap, 1)
}

func TestBuildMergedConfigMap_OverwriteExisting(t *testing.T) {
	pm := newTestPluginManager(t)

	pm.connectionConfigMap = connection.ConnectionConfigMap{
		"conn1": &sdkproto.ConnectionConfig{
			Connection: "conn1",
			Config:     "old",
		},
	}

	newCfg := &sdkproto.ConnectionConfig{
		Connection: "conn1",
		Config:     "new",
	}

	merged := pm.buildMergedConfigMap(newCfg)

	assert.Len(t, merged, 1)
	assert.Equal(t, "new", merged["conn1"].Config)

	// Original map should still have old value
	assert.Equal(t, "old", pm.connectionConfigMap["conn1"].Config)
}

// Test 9: Validate function edge cases
func TestValidateSetConnectionConfigRequest(t *testing.T) {
	// All fields present — should pass
	err := validateSetConnectionConfigRequest(&SetConnectionConfigRequest{
		Connection:      "c",
		PluginShortName: "aws",
		Config:          "x",
		PluginInstance:  "aws",
	})
	assert.NoError(t, err)
}

// Test 10: startConfigAPIServer does nothing when env var is not set
func TestConfigAPIServer_StartDisabledByDefault(t *testing.T) {
	pm := newTestPluginManager(t)

	// Ensure the env var is not set
	t.Setenv("STEAMPIPE_CONFIG_API_PORT", "")

	// Should not panic or start any server
	pm.startConfigAPIServer()
}

func TestConfigAPIServer_StartDisabledWithZero(t *testing.T) {
	pm := newTestPluginManager(t)

	t.Setenv("STEAMPIPE_CONFIG_API_PORT", "0")

	// Should not panic or start any server
	pm.startConfigAPIServer()
}

// Test 13: Concurrent requests for the SAME connection serialize correctly.
// This is the thundering-herd scenario: 20 goroutines all request the same
// connection simultaneously. The per-connection lock ensures they see a
// consistent isNewConnection value (only the first sees true).
func TestConfigAPIServer_SetConnectionConfig_ConcurrentSameConnection(t *testing.T) {
	pm := newTestPluginManager(t)

	var wg sync.WaitGroup
	numGoroutines := 20
	results := make([]int, numGoroutines) // status codes

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			body := makeConfigRequest(t, "same_conn", "hub.steampipe.io/plugins/turbot/aws@latest", "aws",
				`connection "same_conn" { regions = ["us-east-1"] }`, "aws")

			req := httptest.NewRequest(http.MethodPost, "/v1/connection/config", body)
			w := httptest.NewRecorder()

			pm.handleSetConnectionConfig(w, req)
			results[idx] = w.Result().StatusCode
		}(i)
	}

	wg.Wait()

	// All requests should succeed
	for i, code := range results {
		assert.Equal(t, http.StatusOK, code, "goroutine %d failed", i)
	}

	// Only one connection should exist in the map
	pm.mut.RLock()
	defer pm.mut.RUnlock()
	assert.Equal(t, 1, len(pm.connectionConfigMap))
	assert.NotNil(t, pm.connectionConfigMap["same_conn"])
}

// Test 14: getConnLock returns the same mutex for the same connection name
func TestConfigAPIServer_GetConnLock_SameConnection(t *testing.T) {
	pm := newTestPluginManager(t)

	lock1 := pm.getConnLock("conn_a")
	lock2 := pm.getConnLock("conn_a")
	lock3 := pm.getConnLock("conn_b")

	assert.Same(t, lock1, lock2, "same connection should return same mutex")
	assert.NotSame(t, lock1, lock3, "different connections should return different mutexes")
}
