package steampipeconfig

import (
	"sort"
	"strings"
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
	// connection type (expected value: "aggregator")
	Type *string `json:"type,omitempty"  db:"type"`
	// should we create apostgres schema for the connection (expected values: "enable", "disable")
	ImportSchema string `json:"import_schema"  db:"import_schema"`
	// the fully qualified name of the plugin
	Plugin string `json:"plugin,omitempty"  db:"plugin"`
	// the connection state (pending, updating, deleting, error, ready)
	State string `json:"state,omitempty"  db:"state"`
	// error (if there is one - make a pointer to support null)
	ConnectionError *string `json:"error,omitempty" db:"error"`
	// schema mode - static or dynamic
	SchemaMode string `json:"schema_mode,omitempty" db:"schema_mode"`
	// the hash of the connection schema - this is used to determine if a dynamic schema has changed
	SchemaHash string `json:"schema_hash,omitempty" db:"schema_hash"`
	// are the comments set
	CommentsSet bool `json:"comments_set" db:"comments_set"`
	// the creation time of the plugin file
	PluginModTime time.Time `json:"plugin_mod_time" db:"plugin_mod_time"`
	// the update time of the connection
	ConnectionModTime time.Time `json:"connection_mod_time" db:"connection_mod_time"`
	// the matching patterns of child connections (for aggregators)
	Connections []string `json:"connections" db:"connections"`
}

func NewConnectionState(remoteSchema string, connection *modconfig.Connection, creationTime time.Time) *ConnectionState {
	return &ConnectionState{
		Plugin:         remoteSchema,
		ConnectionName: connection.Name,
		PluginModTime:  creationTime,
		State:          constants.ConnectionStateReady,
		Type:           &connection.Type,
		ImportSchema:   connection.ImportSchema,
		Connections:    connection.ConnectionNames,
	}
}

func (d *ConnectionState) Equals(other *ConnectionState) bool {
	if d.Plugin != other.Plugin {
		return false
	}
	if d.GetType() != other.GetType() {
		return false
	}
	if d.ImportSchema != other.ImportSchema {
		return false
	}
	if d.Error() != other.Error() {
		return false
	}

	names := d.Connections
	sort.Strings(names)
	otherNames := other.Connections
	sort.Strings(otherNames)
	if strings.Join(names, ",") != strings.Join(otherNames, "'") {
		return false
	}

	if d.pluginModTimeChanged(other) {
		return false
	}
	// do not look at connection mod time as the mod time for the desired state is not relevant

	return true
}

// allow for sub ms rounding errors when converting from PG
func (d *ConnectionState) pluginModTimeChanged(other *ConnectionState) bool {
	if d.PluginModTime.Sub(other.PluginModTime).Abs() > 1*time.Millisecond {
		return true
	}
	return false
}

func (d *ConnectionState) CanCloneSchema() bool {
	return d.SchemaMode != plugin.SchemaModeDynamic &&
		d.GetType() != modconfig.ConnectionTypeAggregator
}

func (d *ConnectionState) Error() string {
	return typehelpers.SafeString(d.ConnectionError)
}

func (d *ConnectionState) SetError(err string) {
	d.ConnectionError = &err
}

// Loaded returns true if the connection state is 'ready' or 'error'
// Disabled connections are considered as 'loaded'
func (d *ConnectionState) Loaded() bool {
	return d.Disabled() || d.State == constants.ConnectionStateReady || d.State == constants.ConnectionStateError
}

func (d *ConnectionState) Disabled() bool {
	return d.State == constants.ConnectionStateDisabled
}

func (d *ConnectionState) GetType() string {
	return typehelpers.SafeString(d.Type)
}
