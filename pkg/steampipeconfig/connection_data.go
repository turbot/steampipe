package steampipeconfig

import (
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// ConnectionDataStructVersion is used to force refreshing connections
// If we need to force a connection refresh (for example if any of the underlying schema generation code changes),
// updating this version will force all connections to refresh, as the deserialized data will have an old version
var ConnectionDataStructVersion int64 = 20230313

// ConnectionData is a struct containing all details for a connection
// - the plugin name and checksum, the connection config and options
// json tags needed as this is stored in the connection state file
type ConnectionData struct {
	StructVersion int64 `json:"struct_version,omitempty"`
	// the fully qualified name of the plugin
	Plugin string `json:"plugin,omitempty"`
	// the underlying connection object
	Connection *modconfig.Connection `json:"connection,omitempty"`
	// schema mode - static or dynamic
	SchemaMode string `json:"schema_mode,omitempty"`
	// the hash of the connection schema
	SchemaHash string `json:"schema_hash,omitempty"`
	// the creation time of the plugin file (only used for local plugins)
	ModTime time.Time `json:"mod_time"`
	// loaded is false if the plugin failed to load
	Loaded bool `json:"loaded"`
	// error to be populated if we failed to start/load plugin
	Error string `json:"error,omitempty"`
}

func NewConnectionData(remoteSchema string, connection *modconfig.Connection, creationTime time.Time) *ConnectionData {
	return &ConnectionData{
		StructVersion: ConnectionDataStructVersion,
		Plugin:        remoteSchema,
		Connection:    connection,
		ModTime:       creationTime,
		Loaded:        true,
	}
}

// IsValid checks whether the struct was correctly deserialized,
// by checking if the StructVersion is populated
func (d *ConnectionData) IsValid() bool {
	return d.StructVersion > 0
}

func (d *ConnectionData) Equals(other *ConnectionData) bool {
	if d.Connection == nil || other.Connection == nil {
		// if either object has a nil Connection, then it may be data from an old connection state file
		// return false, so that connections get refreshed and this file gets written in the new format in the process
		return false
	}

	return d.Plugin == other.Plugin &&
		d.Connection.Equals(other.Connection) &&
		d.ModTime.Equal(other.ModTime) &&
		d.Connection.Equals(other.Connection)
}

func (d *ConnectionData) CanCloneSchema() bool {
	return d.SchemaMode != plugin.SchemaModeDynamic &&
		d.Connection.Type != modconfig.ConnectionTypeAggregator
}
