package steampipeconfig

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/turbot/steampipe/filepaths"
	"github.com/turbot/steampipe/migrate"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ConnectionDataStructVersion is used to force refreshing connections
// If we need to force a connection refresh (for example if any of the underlying schema generation code changes),
// updating this version will force all connections to refresh, as the deserialized data will have an old version
var ConnectionDataStructVersion int64 = 20211125

// LegacyConnectionData is the legacy connection data struct, which was used in the legacy
// connection state file
type LegacyConnectionData struct {
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
}

// IsValid checks whether the struct was correctly deserialized,
// by checking if the StructVersion is populated
func (s ConnectionData) IsValid() bool {
	return s.StructVersion > 0
}

func (s *ConnectionData) MigrateFrom(prev interface{}) migrate.Migrateable {
	legacyState := prev.(LegacyConnectionData)
	s.StructVersion = ConnectionDataStructVersion
	s.Plugin = legacyState.Plugin
	s.SchemaMode = legacyState.SchemaMode
	s.SchemaHash = legacyState.SchemaHash
	s.ModTime = legacyState.ModTime
	s.Connection = legacyState.Connection

	return s
}

func LegacyStateFilePath() string {
	return filepath.Join(filepaths.EnsureInternalDir(), "connection.json")
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

func (f *ConnectionData) Save() error {
	versionFilePath := filepaths.ConnectionStatePath()
	return f.write(versionFilePath)
}

func (f *ConnectionData) write(path string) error {
	versionFileJSON, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		log.Println("[ERROR]", "Error while writing state file", err)
		return err
	}
	return os.WriteFile(path, versionFileJSON, 0644)
}
