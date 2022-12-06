## Modfile Parsing
Modfile parsing and decoding is executed before the remainder of the mod resources are parsed.
This is necessary as we need to identify mod dependencies.

**This means we DO NOT support hcl references within the modfile defintion**

The exception to this is when passing args to dependent mods. These are parsed separately at the end os the workspace parse process

##Database setup and Initialisation

DB Installation is ensured by calling `EnsureDBInstalled`

### Overview
If the database `IsInstalled`, it calls `prepareDb`

If not installed, the db is installed, migrating public schem data if there is a major version update

### Details
#### IsInstalled
This function determines if the database is installed by:
- looking for the initDb, postgres and fdw binaries
- looking for the fdw control and sql file

If any of these are missing the database is deemed not installed and a full installation occurs

#### prepareDb
This function:
- checks if the installed db version has the correct ImageRef. In other words, has the database package changed, without the Postgres version changing. If so, it installs the new database package (and FDW), retaining the existing data
- checks if the correct FDW version is installed - if not it installs it
- checks if the database is initialised (by testing whether pg_hba.conf exists) and if not, initialise it *TODO* identify when this can occur


#### Database Installation
If `IsInstalled` returns false a full db installation is carried out.
- first verify if a service is running. If so, display an error and return
- download and install the db files
- if this is a Major version update, use pg_dump to backup the public schema
- install the FDW

NOTE: if a backup was taken it is restored by `restoreBackup` which is called from `RefreshConnectionAndSearchPaths` -
MOVE THIS: https://github.com/turbot/steampipe/issues/2037
