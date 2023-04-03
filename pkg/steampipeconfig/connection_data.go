package steampipeconfig

import (
	"log"
	"time"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// ConnectionDataStructVersion is used to force refreshing connections
// If we need to force a connection refresh (for example if any of the underlying schema generation code changes),
// updating this version will force all connections to refresh, as the deserialized data will have an old version
var ConnectionDataStructVersion int64 = 20230330

// ConnectionData is a struct containing all details for a connection
// - the plugin name and checksum, the connection config and options
// json tags needed as this is stored in the connection state file
type ConnectionData struct {
	StructVersion int64 `json:"struct_version,omitempty" db:"-"`
	// the connection name
	ConnectionName string `json:"connection,omitempty"  db:"name"`
	// the connection object
	Connection *modconfig.Connection `json:"connection,omitempty"  db:"-"`
	// the fully qualified name of the plugin
	Plugin string `json:"plugin,omitempty"  db:"plugin"`
	// the connection state (pending, updating, deleting, error, ready)
	ConnectionState string `json:"state,omitempty"  db:"state"`
	// error (if there is one - make a pointer to supprt null)
	ConnectionError *string `json:"error,omitempty" db:"error"`
	// schema mode - static or dynamic
	SchemaMode string `json:"schema_mode,omitempty" db:"schema_mode"`
	// the hash of the connection schema - this is used to determine if a dynamic schema has changed
	SchemaHash string `json:"schema_hash,omitempty" db:"schema_hash"`
	// the creation time of the plugin file
	PluginModTime time.Time `json:"plugin_mod_time" db:"plugin_mod_time"`
	//// the update time of the connection
	//ConnectionModTime time.Time `json:"conneciton_mod_time"`
	// loaded is false if the plugin failed to load
	//Loaded bool `json:"loaded"`
}

func NewConnectionData(remoteSchema string, connection *modconfig.Connection, creationTime time.Time) *ConnectionData {
	return &ConnectionData{
		StructVersion:   ConnectionDataStructVersion,
		Plugin:          remoteSchema,
		ConnectionName:  connection.Name,
		Connection:      connection,
		PluginModTime:   creationTime,
		ConnectionState: constants.ConnectionStateReady,
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

	// TODO KAI remove debug version

	if d.Plugin != other.Plugin {
		return false
	}
	if d.ConnectionState != other.ConnectionState {
		return false
	}
	if d.Error() != other.Error() {
		return false
	}
	if !d.Connection.Equals(other.Connection) {
		return false
	}
	// allow for sub ms rounding errors when converting from PG
	if d.PluginModTime.Sub(other.PluginModTime).Abs() > 1*time.Millisecond {
		a := d.PluginModTime.Sub(other.PluginModTime)
		log.Printf("[WARN] %v", a)
		return false
	}
	//d.ConnectionModTime.Equal(other.ConnectionModTime) return false
	if !d.Connection.Equals(other.Connection) {
		return false
	}
	//}	return d.Plugin == other.Plugin &&
	//		d.ConnectionState == other.ConnectionState &&
	//		d.Error() == other.Error() &&
	//		d.SchemaMode == other.SchemaMode &&
	//		d.Connection.Equals(other.Connection) &&
	//		d.PluginModTime.Equal(other.PluginModTime) &&
	//		//d.ConnectionModTime.Equal(other.ConnectionModTime) &&
	//		d.Connection.Equals(other.Connection)
	return true
}

func (d *ConnectionData) CanCloneSchema() bool {
	return d.SchemaMode != plugin.SchemaModeDynamic &&
		d.Connection.Type != modconfig.ConnectionTypeAggregator
}

func (d *ConnectionData) Error() string {
	return typehelpers.SafeString(d.ConnectionError)

}
func (d *ConnectionData) SetError(err string) {
	d.ConnectionError = &err
}
