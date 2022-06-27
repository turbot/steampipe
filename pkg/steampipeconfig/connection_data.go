package steampipeconfig

import (
	"time"

	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// ConnectionDataStructVersion is used to force refreshing connections
// If we need to force a connection refresh (for example if any of the underlying schema generation code changes),
// updating this version will force all connections to refresh, as the deserialized data will have an old version
var ConnectionDataStructVersion int64 = 20211125

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

	// legacy properties included for backwards compatibility with v0.13
	LegacyPlugin     string                `json:"Plugin,omitempty"`
	LegacyConnection *modconfig.Connection `json:"Connection,omitempty"`
	LegacySchemaMode string                `json:"SchemaMode,omitempty"`
	LegacySchemaHash string                `json:"SchemaHash,omitempty"`
	LegacyModTime    time.Time             `json:"ModTime,omitempty"`
}

// IsValid checks whether the struct was correctly deserialized,
// by checking if the StructVersion is populated
func (s *ConnectionData) IsValid() bool {
	return s.StructVersion > 0
}

// MigrateLegacy migrates the legacy properties into new properties
func (s *ConnectionData) MigrateLegacy() {
	s.StructVersion = ConnectionDataStructVersion
	s.Plugin = s.LegacyPlugin
	s.SchemaMode = s.LegacySchemaMode
	s.SchemaHash = s.LegacySchemaHash
	s.ModTime = s.LegacyModTime
	s.Connection = s.LegacyConnection
	s.Connection.MigrateLegacy()
}

// MaintainLegacy keeps the values of the legacy properties intact while
// refreshing connections
func (s *ConnectionData) MaintainLegacy() {
	s.LegacyPlugin = s.Plugin
	s.LegacySchemaMode = s.SchemaMode
	s.LegacySchemaHash = s.SchemaHash
	s.LegacyModTime = s.ModTime
	s.LegacyConnection = s.Connection
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
