package pluginmanager_service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	sdkproto "github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe/v2/pkg/connection"
	"github.com/turbot/steampipe/v2/pkg/db/db_common"
)

// SetConnectionConfigRequest is the JSON request body for the config API endpoint.
type SetConnectionConfigRequest struct {
	Connection      string `json:"connection"`
	Plugin          string `json:"plugin"`
	PluginShortName string `json:"plugin_short_name"`
	Config          string `json:"config"`
	PluginInstance  string `json:"plugin_instance"`
}

// SetConnectionConfigResponse is the JSON response from the config API endpoint.
type SetConnectionConfigResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// startConfigAPIServer starts the HTTP config API server if STEAMPIPE_CONFIG_API_PORT is set.
// The server runs in a goroutine and provides a synchronous endpoint for updating
// connection configs, bypassing the file watcher.
func (m *PluginManager) startConfigAPIServer() {
	port := os.Getenv("STEAMPIPE_CONFIG_API_PORT")
	if port == "" || port == "0" {
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/connection/config", m.handleSetConnectionConfig)

	addr := fmt.Sprintf("127.0.0.1:%s", port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		log.Printf("[INFO] Config API server listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[WARN] Config API server error: %s", err.Error())
		}
	}()
}

// getConnLock returns the per-connection mutex for the given connection name,
// creating one if it doesn't exist. This serializes concurrent requests for the
// same connection so that ensureConnectionSchema completes before any concurrent
// request can return success. Different connections are fully parallel.
func (m *PluginManager) getConnLock(connectionName string) *sync.Mutex {
	m.connLocksMu.Lock()
	mu, ok := m.connLocks[connectionName]
	if !ok {
		mu = &sync.Mutex{}
		m.connLocks[connectionName] = mu
	}
	m.connLocksMu.Unlock()
	return mu
}

// handleSetConnectionConfig handles POST /v1/connection/config.
// It updates a single connection's config in the plugin manager synchronously,
// sending the update to the running plugin via gRPC before returning.
//
// Per-connection locking ensures that when multiple concurrent requests arrive
// for the same connection, only the first creates the schema — the rest wait
// until it's done, then proceed (as credential-rotation updates, not new connections).
func (m *PluginManager) handleSetConnectionConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SetConnectionConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeConfigAPIResponse(w, http.StatusBadRequest, SetConnectionConfigResponse{
			Error: fmt.Sprintf("invalid JSON: %s", err.Error()),
		})
		return
	}

	if err := validateSetConnectionConfigRequest(&req); err != nil {
		writeConfigAPIResponse(w, http.StatusBadRequest, SetConnectionConfigResponse{
			Error: err.Error(),
		})
		return
	}

	// Acquire per-connection lock. This serializes the entire config-update +
	// schema-creation sequence for the same connection name. Without this, 30
	// concurrent requests all see isNewConnection=true (first) or false (rest),
	// but the "rest" return 200 before the first thread's ensureConnectionSchema
	// has finished, causing "relation does not exist" errors.
	connLock := m.getConnLock(req.Connection)
	connLock.Lock()
	defer connLock.Unlock()

	// Build the new ConnectionConfig proto
	newConfig := &sdkproto.ConnectionConfig{
		Connection:      req.Connection,
		Plugin:          req.Plugin,
		PluginShortName: req.PluginShortName,
		Config:          req.Config,
		PluginInstance:  req.PluginInstance,
	}

	// Acquire the PluginManager state lock and update the config map
	m.mut.Lock()

	// Track whether this is a new connection (needs schema creation) or an update (creds only)
	_, isNewConnection := m.connectionConfigMap[req.Connection]
	isNewConnection = !isNewConnection

	// Resolve PluginInstance and Plugin to match the full image refs used as keys
	// in m.plugins (e.g., "hub.steampipe.io/plugins/turbot/aws@latest"), not short
	// names like "aws". For existing connections, inherit from the current config.
	// For new connections, inherit from any existing connection using the same plugin.
	if existing, ok := m.connectionConfigMap[req.Connection]; ok {
		if newConfig.PluginInstance != existing.PluginInstance {
			newConfig.PluginInstance = existing.PluginInstance
		}
		if newConfig.Plugin == "" {
			newConfig.Plugin = existing.Plugin
		}
	} else {
		// New connection: resolve PluginInstance from any existing connection
		// that uses the same plugin, so the FDW can find the running plugin process.
		for _, existing := range m.connectionConfigMap {
			if existing.PluginShortName == req.PluginShortName {
				newConfig.PluginInstance = existing.PluginInstance
				if newConfig.Plugin == "" {
					newConfig.Plugin = existing.Plugin
				}
				break
			}
		}
	}

	newConfigMap := m.buildMergedConfigMap(newConfig)
	err := m.handleConnectionConfigChanges(context.Background(), newConfigMap)
	m.mut.Unlock()

	if err != nil {
		log.Printf("[WARN] Config API: handleConnectionConfigChanges failed for connection '%s': %s", req.Connection, err.Error())
		writeConfigAPIResponse(w, http.StatusInternalServerError, SetConnectionConfigResponse{
			Error: fmt.Sprintf("failed to update connection config: %s", err.Error()),
		})
		return
	}

	// For new connections, create the FDW schema in postgres so the connection
	// is immediately queryable. For existing connections (credential rotation),
	// the schema already exists and only the credentials needed updating.
	// This runs under connLock, so concurrent requests for the same connection
	// wait until schema creation is complete before proceeding.
	if isNewConnection && m.pool != nil {
		if err := m.ensureConnectionSchema(context.Background(), req.Connection, req.PluginShortName); err != nil {
			log.Printf("[WARN] Config API: schema creation failed for connection '%s': %s", req.Connection, err.Error())
			writeConfigAPIResponse(w, http.StatusInternalServerError, SetConnectionConfigResponse{
				Error: fmt.Sprintf("connection config updated but schema creation failed: %s", err.Error()),
			})
			return
		}
	}

	log.Printf("[INFO] Config API: updated connection '%s' (plugin: %s)", req.Connection, req.PluginShortName)
	writeConfigAPIResponse(w, http.StatusOK, SetConnectionConfigResponse{Success: true})
}

// buildMergedConfigMap creates a new ConnectionConfigMap by copying the existing map
// and adding or updating the given connection config.
func (m *PluginManager) buildMergedConfigMap(config *sdkproto.ConnectionConfig) connection.ConnectionConfigMap {
	newMap := make(connection.ConnectionConfigMap, len(m.connectionConfigMap)+1)
	for k, v := range m.connectionConfigMap {
		newMap[k] = v
	}
	newMap[config.Connection] = config
	return newMap
}

func validateSetConnectionConfigRequest(req *SetConnectionConfigRequest) error {
	if req.Connection == "" {
		return fmt.Errorf("'connection' is required")
	}
	if req.PluginShortName == "" {
		return fmt.Errorf("'plugin_short_name' is required")
	}
	if req.Config == "" {
		return fmt.Errorf("'config' is required")
	}
	if req.PluginInstance == "" {
		return fmt.Errorf("'plugin_instance' is required")
	}
	return nil
}

// ensureConnectionSchema creates the FDW schema for a new connection.
// The plugin manager config was already updated in-memory via
// handleConnectionConfigChanges(), so the FDW import path can resolve this
// connection directly through plugin manager gRPC without any disk config files.
func (m *PluginManager) ensureConnectionSchema(ctx context.Context, connectionName, pluginShortName string) error {
	// Create the FDW schema in postgres
	sql := db_common.GetUpdateConnectionQuery(connectionName, pluginShortName)
	log.Printf("[INFO] Config API: creating schema for connection '%s' (plugin schema: %s)", connectionName, pluginShortName)

	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire db connection: %w", err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to create connection schema: %w", err)
	}

	// Notify listeners that the schema has changed
	if notifyErr := m.SendPostgresSchemaNotification(ctx); notifyErr != nil {
		log.Printf("[WARN] Config API: failed to send schema notification: %s", notifyErr.Error())
	}

	return nil
}

func writeConfigAPIResponse(w http.ResponseWriter, status int, resp SetConnectionConfigResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
