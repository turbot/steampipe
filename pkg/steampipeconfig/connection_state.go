package steampipeconfig

import (
	"time"

	typehelpers "github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
)

// ConnectionState is a struct containing all details for a connection
// - the plugin name and checksum, the connection config and options
// json tags needed as this is stored in the connection state file
type ConnectionState struct {
	// the connection name
	ConnectionName string `json:"connection,omitempty"  db:"name"`
	// the connection object
	Connection *modconfig.Connection `json:"connection,omitempty"  db:"-"`
	// the fully qualified name of the plugin
	Plugin string `json:"plugin,omitempty"  db:"plugin"`
	// the connection state (pending, updating, deleting, error, ready)
	State string `json:"state,omitempty"  db:"state"`
	// error (if there is one - make a pointer to supprt null)
	ConnectionError *string `json:"error,omitempty" db:"error"`
	// schema mode - static or dynamic
	SchemaMode string `json:"schema_mode,omitempty" db:"schema_mode"`
	// the hash of the connection schema - this is used to determine if a dynamic schema has changed
	SchemaHash string `json:"schema_hash,omitempty" db:"schema_hash"`
	// the creation time of the plugin file
	PluginModTime time.Time `json:"plugin_mod_time" db:"plugin_mod_time"`
	// the update time of the connection
	ConnectionModTime time.Time `json:"connection_mod_time" db:"connection_mod_time"`
}

func NewConnectionData(remoteSchema string, connection *modconfig.Connection, creationTime time.Time) *ConnectionState {
	return &ConnectionState{
		Plugin:         remoteSchema,
		ConnectionName: connection.Name,
		Connection:     connection,
		PluginModTime:  creationTime,
		State:          constants.ConnectionStateReady,
	}
}

func (d *ConnectionState) Equals(other *ConnectionState) bool {
	if d.Connection == nil || other.Connection == nil {
		// if either object has a nil Connection, then it may be data from an old connection state file
		// return false, so that connections get refreshed and this file gets written in the new format in the process
		return false
	}
	if d.Plugin != other.Plugin {
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
		return false
	}
	//d.ConnectionModTime.Equal(other.ConnectionModTime) return false
	if !d.Connection.Equals(other.Connection) {
		return false
	}

	return true
}

func (d *ConnectionState) CanCloneSchema() bool {
	return d.SchemaMode != plugin.SchemaModeDynamic &&
		d.Connection.Type != modconfig.ConnectionTypeAggregator
}

func (d *ConnectionState) Error() string {
	return typehelpers.SafeString(d.ConnectionError)

}
func (d *ConnectionState) SetError(err string) {
	d.ConnectionError = &err
}

// Loaded returns true if the connection state is 'ready' or 'error'
func (d *ConnectionState) Loaded() bool {
	return d.State == constants.ConnectionStateReady || d.State == constants.ConnectionStateError
}
