package constants

import "github.com/turbot/steampipe/schema"

// dbClient constants
// TODO these should be configuration settings

// Arrays cannot be constants, so do the next most convenient thing
var DatabaseListenAddresses = []string{"localhost", "127.0.0.1"}

const (
	DatabaseHost      = "localhost"
	DatabasePort      = 9193
	DatabaseSuperUser = "root"
	DatabaseUser      = "steampipe"
	DatabaseName      = "steampipe"
)

// constants for installing db and fdw images
const (
	DatabaseVersion = "12.1.0"
	FdwVersion      = "0.0.23-beta.0"

	// The 12.1.0 image uses the older jar format 12.1.0-v2 is the same version of postgres,
	// just packaged as gzipped tar files (consistent with oras, faster to unzip).  Once everyone is
	// on a newer build, we can delete the old image move the 12.1.0 tag to the new image, and
	// change this back for consistency
	//DefaultEmbeddedPostgresImage = "us-docker.pkg.dev/steampipe/steampipe/db:" + DatabaseVersion
	DefaultEmbeddedPostgresImage = "us-docker.pkg.dev/steampipe/steampipe/db:12.1.0-v2"
	DefaultFdwImage              = "us-docker.pkg.dev/steampipe/steampipe/fdw:" + FdwVersion
)

// FunctionSchema :: schema container for all steampipe helper functions
const FunctionSchema = "internal"

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
