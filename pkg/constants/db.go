package constants

import (
	"fmt"
)

// Client constants
const (
	// MaxParallelClientInits is the number of clients to initialize in parallel
	// if we start initializing all clients together, it leads to bad performance on all
	MaxParallelClientInits = 3

	// MaxBackups is the maximum number of backups that will be retained
	MaxBackups = 100
)

const (
	DatabaseDefaultListenAddresses   = "localhost"
	DatabaseDefaultPort              = 9193
	DatabaseDefaultCheckQueryTimeout = 240
	DatabaseSuperUser                = "root"
	DatabaseUser                     = "steampipe"
	DatabaseName                     = "steampipe"
	DatabaseUsersRole                = "steampipe_users"
	DefaultMaxConnections            = 10
)

// constants for installing db and fdw images
const (
	DatabaseVersion = "14.19.0"
	FdwVersion      = "2.1.4"

	// PostgresImageRef is the OCI Image ref for the database binaries
	PostgresImageRef    = "ghcr.io/turbot/steampipe/db:14.19.0"
	PostgresImageDigest = "sha256:84264ef41853178707bccb091f5450c22e835f8a98f9961592c75690321093d9"

	FdwImageRef       = "ghcr.io/turbot/steampipe/fdw:" + FdwVersion
	FdwBinaryFileName = "steampipe_postgres_fdw.so"
)

// schema names
const (

	// legacy schema names
	// these are schema names which were used previously
	// but are not relevant anymore and need to be dropped
	LegacyInternalSchema = "internal"

	// InternalSchema is the schema container for all steampipe helper functions, and connection state table
	// also used to send commands to the FDW
	InternalSchema = "steampipe_internal"

	// ServerSettingsTable is the table used to store steampipe service configuration
	ServerSettingsTable = "steampipe_server_settings"

	// RateLimiterDefinitionTable is the table used to store rate limiters defined in the config
	RateLimiterDefinitionTable = "steampipe_plugin_limiter"
	// PluginInstanceTable is the table used to store plugin configs
	PluginInstanceTable = "steampipe_plugin"
	PluginColumnTable   = "steampipe_plugin_column"

	// LegacyConnectionStateTable is the table used to store steampipe connection state
	LegacyConnectionStateTable       = "steampipe_connection_state"
	ConnectionTable                  = "steampipe_connection"
	ConnectionStatePending           = "pending"
	ConnectionStatePendingIncomplete = "incomplete"
	ConnectionStateReady             = "ready"
	ConnectionStateUpdating          = "updating"
	ConnectionStateDeleting          = "deleting"
	ConnectionStateDisabled          = "disabled"
	ConnectionStateError             = "error"

	// foreign tables in internal schema
	ForeignTableScanMetadataSummary       = "steampipe_scan_metadata_summary"
	ForeignTableScanMetadata              = "steampipe_scan_metadata"
	ForeignTableSettings                  = "steampipe_settings"
	ForeignTableSettingsKeyColumn         = "name"
	ForeignTableSettingsValueColumn       = "value"
	ForeignTableSettingsCacheKey          = "cache"
	ForeignTableSettingsCacheTtlKey       = "cache_ttl"
	ForeignTableSettingsCacheClearTimeKey = "cache_clear_time"

	FunctionCacheSet             = "meta_cache"
	FunctionConnectionCacheClear = "meta_connection_cache_clear"
	FunctionCacheSetTtl          = "meta_cache_ttl"

	// legacy
	LegacyCommandSchema = "steampipe_command"

	LegacyCommandTableCache                = "cache"
	LegacyCommandTableCacheOperationColumn = "operation"
	LegacyCommandCacheOn                   = "cache_on"
	LegacyCommandCacheOff                  = "cache_off"
	LegacyCommandCacheClear                = "cache_clear"

	LegacyCommandTableScanMetadata = "scan_metadata"
)

// ConnectionStates is a handy array of all states
var ConnectionStates = []string{
	LegacyConnectionStateTable,
	ConnectionStatePending,
	ConnectionStateReady,
	ConnectionStateUpdating,
	ConnectionStateDeleting,
	ConnectionStateError,
}

var ReservedConnectionNames = []string{
	"public",
}

const ReservedConnectionNamePrefix = "steampipe_"

// introspection table names
const (
	IntrospectionTableQuery              = "steampipe_query"
	IntrospectionTableControl            = "steampipe_control"
	IntrospectionTableBenchmark          = "steampipe_benchmark"
	IntrospectionTableMod                = "steampipe_mod"
	IntrospectionTableDashboard          = "steampipe_dashboard"
	IntrospectionTableDashboardContainer = "steampipe_dashboard_container"
	IntrospectionTableDashboardCard      = "steampipe_dashboard_card"
	IntrospectionTableDashboardChart     = "steampipe_dashboard_chart"
	IntrospectionTableDashboardFlow      = "steampipe_dashboard_flow"
	IntrospectionTableDashboardGraph     = "steampipe_dashboard_graph"
	IntrospectionTableDashboardHierarchy = "steampipe_dashboard_hierarchy"
	IntrospectionTableDashboardImage     = "steampipe_dashboard_image"
	IntrospectionTableDashboardInput     = "steampipe_dashboard_input"
	IntrospectionTableDashboardTable     = "steampipe_dashboard_table"
	IntrospectionTableDashboardText      = "steampipe_dashboard_text"
	IntrospectionTableVariable           = "steampipe_variable"
	IntrospectionTableReference          = "steampipe_reference"
)

const (
	RuntimeParamsKeyApplicationName = "application_name"
)

// Invoker is a pseudoEnum for the command/operation which starts the service
type Invoker string

const (
	// InvokerService is set when invoked by `service start`
	InvokerService Invoker = "service"
	// InvokerQuery is set when invoked by query command
	InvokerQuery = "query"
	// InvokerCheck is set when invoked by check command
	InvokerCheck = "check"
	// InvokerPlugin is set when invoked by a plugin command
	InvokerPlugin = "plugin"
	// InvokerDashboard is set when invoked by dashboard command
	InvokerDashboard = "dashboard"
	// InvokerConnectionWatcher is set when invoked by the connection watcher process
	InvokerConnectionWatcher = "connection-watcher"
)

// IsValid is a validator for Invoker known values
func (i Invoker) IsValid() error {
	switch i {
	case InvokerService, InvokerQuery, InvokerCheck, InvokerPlugin, InvokerDashboard:
		return nil
	}
	return fmt.Errorf("invalid invoker. Can be one of '%v', '%v', '%v', '%v' or '%v' ", InvokerService, InvokerQuery, InvokerPlugin, InvokerCheck, InvokerDashboard)
}
