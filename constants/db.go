package constants

import (
	"fmt"

	"github.com/turbot/steampipe/schema"
)

// dbClient constants
// TODO these should be configuration settings

// DatabaseListenAddresses is an arrays is listen addresses which Steampipe accepts
var DatabaseListenAddresses = []string{"localhost", "127.0.0.1"}

const (
	DatabaseHost        = "localhost"
	DatabaseDefaultPort = 9193
	DatabaseSuperUser   = "root"
	DatabaseUser        = "steampipe"
	DatabaseName        = "steampipe"
)

// constants for installing db and fdw images
const (
	DatabaseVersion = "12.1.0"
	FdwVersion      = "0.2.3"

	// DefaultEmbeddedPostgresImage :: The 12.1.0 image uses the older jar format 12.1.0-v2 is the same version of postgres,
	// just packaged as gzipped tar files (consistent with oras, faster to unzip).  Once everyone is
	// on a newer build, we can delete the old image move the 12.1.0 tag to the new image, and
	// change this back for consistency
	//DefaultEmbeddedPostgresImage = "us-docker.pkg.dev/steampipe/steampipe/db:" + DatabaseVersion
	DefaultEmbeddedPostgresImage = "us-docker.pkg.dev/steampipe/steampipe/db:12.1.0-v2"
	DefaultFdwImage              = "us-docker.pkg.dev/steampipe/steampipe/fdw:" + FdwVersion
)

// schema names
const (
	// FunctionSchema :: schema container for all steampipe helper functions
	FunctionSchema = "internal"

	// CommandSchema is the schema which is used to send commands to the FDW
	CommandSchema               = "steampipe_command"
	CacheCommandTable           = "cache"
	CacheCommandOperationColumn = "operation"
	CommandCacheOn              = "cache_on"
	CommandCacheOff             = "cache_off"
	CommandCacheClear           = "cache_clear"
)

// Functions :: a list of SQLFunc objects that are installed in the db 'internal' schema startup
var Functions = []schema.SQLFunc{
	{
		Name:     "glob",
		Params:   map[string]string{"input_glob": "text"},
		Returns:  "text",
		Language: "plpgsql",
		Body: `
declare
	output_pattern text;
begin
	output_pattern = replace(input_glob, '*', '%');
	output_pattern = replace(output_pattern, '?', '_');
	return output_pattern;
end;
`,
	},
}

var ReservedConnectionNames = []string{
	"public",
	FunctionSchema,
}

// reflection table names
const (
	ReflectionTableQuery     = "steampipe_query"
	ReflectionTableControl   = "steampipe_control"
	ReflectionTableBenchmark = "steampipe_benchmark"
	ReflectionTableMod       = "steampipe_mod"
)

func ReflectionTableNames() []string {
	return []string{ReflectionTableControl, ReflectionTableBenchmark, ReflectionTableQuery, ReflectionTableMod}
}

// Invoker is a pseudoEnum for the command/operation which starts the service
type Invoker string

const (
	// InvokerService is set when invoked by `service start`
	InvokerService Invoker = "service"
	// InvokerQuery is set when invoked by query command
	InvokerQuery = "query"
	// InvokerCheck is set when invoked by check command
	InvokerCheck = "check"
	// InvokerInstaller is set when when invoked by the installer
	InvokerInstaller = "installer"
	// InvokerPlugin is set when invoked by a plugin command
	InvokerPlugin = "plugin"
	// InvokerReport is set when invoked by report command
	InvokerReport = "report"
)

// IsValid is a validator for Invoker known values
func (i Invoker) IsValid() error {
	switch i {
	case InvokerService, InvokerQuery, InvokerCheck, InvokerInstaller, InvokerPlugin, InvokerReport:
		return nil
	}
	return fmt.Errorf("Invalid invoker. Can be one of '%v', '%v', '%v', '%v', '%v' or '%v' ", InvokerService, InvokerQuery, InvokerInstaller, InvokerPlugin, InvokerCheck, InvokerReport)
}
