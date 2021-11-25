package steampipeconfig

import (
	"time"

	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ConnectionDataStructVersion is used to force refreshing connections
// If we need to force a connection refresh (for example if any of the underlying schema generation code changes),
// updating this version will force all connections to refresh, as the deserialized data will have an old version
var ConnectionDataStructVersion int64 = 20211125

// ConnectionData is a struct containing all details for a connection
// - the plugin name and checksum, the connection config and options
// json tags needed as this is stored in the connection state file
type ConnectionData struct {
	StructVersion int64
	// the fully qualified name of the plugin
	Plugin string
	// the underlying connection object
	Connection *modconfig.Connection
	// schema mode - static or dynamic
	SchemaMode string `json:"SchemaMode,omitempty"`
	// the hash of the connection schema
	SchemaHash string `json:"SchemaHash,omitempty"`
	// the creation time of the plugin file (only used for local plugins)
	ModTime time.Time
}

func NewConnectionData(remoteSchema string, connection *modconfig.Connection, creationTime time.Time) *ConnectionData {
	return &ConnectionData{
		StructVersion: ConnectionDataStructVersion,
		Plugin:        remoteSchema,
		Connection:    connection,
		ModTime:       creationTime,
	}
}

func (p *ConnectionData) Equals(other *ConnectionData) bool {
	if p.Connection == nil || other.Connection == nil {
		// if either object has a nil Connection, then it may be data from an old connection state file
		// return false, so that connections get refreshed and this file gets written in the new format in the process
		return false
	}

	return p.Plugin == other.Plugin &&
		p.ModTime.Equal(other.ModTime) &&
		p.Connection.Equals(other.Connection)
}
