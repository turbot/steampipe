package constants

import (
	"fmt"

	"github.com/turbot/steampipe/schema"
)

// dbClient constants
// TODO these should be configuration settings

// Arrays cannot be constants, so do the next most convenient thing
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
	FdwVersion      = "0.1.2"

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
	ReflectionTableQuery      = "steampipe_query"
	ReflectionTableControl    = "steampipe_control"
	ReflectionTableBenchmark  = "steampipe_benchmark"
	ReflectionTableMod        = "steampipe_mod"
	ReflectionTableConnection = "steampipe_connection"
)

func ReflectionTableNames() []string {
	return []string{ReflectionTableControl, ReflectionTableBenchmark, ReflectionTableQuery, ReflectionTableMod, ReflectionTableConnection}
}

// Invoker :: pseudoEnum for what starts the service
type Invoker string

const (
	// InvokerService :: Invoker - when invoked by `service start`
	InvokerService Invoker = "service"
	// InvokerQuery :: Invoker - when invoked by `query`
	InvokerQuery = "query"
	// InvokerCheck :: Invoker - when invoked by `check`
	InvokerCheck = "check"
	// InvokerInstaller :: Invoker - when invoked by the `installer`
	InvokerInstaller = "installer"
	// InvokerPlugin :: Invoker - when invoked by the `pluginmanager`
	InvokerPlugin = "plugin"
	// InvokerReport :: Invoker - when invoked by `report`
	InvokerReport = "report"
)

// TODO - this is a bit naff

// IsValid :: validator for Invoker known values
func (slt Invoker) IsValid() error {
	switch slt {
	case InvokerService, InvokerQuery, InvokerCheck, InvokerInstaller, InvokerPlugin, InvokerReport:
		return nil
	}
	return fmt.Errorf("Invalid invoker. Can be one of '%v', '%v', '%v', '%v', '%v' or '%v' ", InvokerService, InvokerQuery, InvokerInstaller, InvokerPlugin, InvokerCheck, InvokerReport)
}
