package constants

var MinimalPgHbaContent string = `
hostssl all root samehost trust
host all root samehost trust
`

// PgHbaTemplate is to be formatted with two variables:
// 		* databaseName
//		* username
//
// Example:
//		fmt.Sprintf(template, datName, username)
var PgHbaTemplate string = `
# PostgreSQL Client Authentication Configuration File
# ===================================================
#
# STEAMPIPE
#
# The root user is assumed by steampipe to manage the database configuration.
# Access is not granted to users of steampipe.
#
# The configuration is:
# * Access is restricted to samehost
# * Future - access via SSL only (remove host line)
#
hostssl all root samehost trust
host    all root samehost trust

# All user queries (steampipe query, steampipe service etc.) are run as the
# steampipe user. The steampipe user is restricted in access to the steampipe
# database, and further restricted by permissions to only read from steampipe
# managed schemas. Write access is allowed to the public schema in the
# steampipe database.
#
# The configuration is:
# * Access from samehost does not require a password (trust)
# * Access from any other host does require a password
# * Future - access via SSL only (remove host line)
#
hostssl %[1]s %[2]s samehost trust
host    %[1]s %[2]s samehost trust
hostssl %[1]s %[2]s all scram-sha-256
host    %[1]s %[2]s all scram-sha-256
`
