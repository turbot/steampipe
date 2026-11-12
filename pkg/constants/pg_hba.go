package constants

// PgHbaHeader starts every pg_hba.conf that Steampipe writes. A checksum line follows it.
// Steampipe recognises a file that it wrote by the first line, and reads the checksum only
// from a file that starts with every line of the header followed by the checksum prefix: a
// file with any other header counts as edited. Changing any header line must keep accepting
// the old header, or every file written with it counts as edited.
const PgHbaHeader = `# Written by Steampipe each time the database starts. DO NOT EDIT.
# Put your own rules in pg_hba.conf.d/<name>.conf. They come first, so they win.
# To manage this file yourself, delete these header lines. Steampipe then does not change it,
# until a Postgres major version upgrade starts a new data folder with Steampipe's file.
`

// PgHbaChecksumPrefix starts the line after PgHbaHeader. The rest of the line is the hex
// SHA-256 of every line after it, so Steampipe can tell that someone edited its file.
const PgHbaChecksumPrefix = "# checksum: sha256:"

// PgHbaSteampipeRules are Steampipe's own rules. They are written after every rule
// in pg_hba.conf.d.
//
// The database column is "all": at install, CONNECT is revoked from PUBLIC on the postgres
// and template1 databases, so the steampipe user can reach only its own database. A database
// that someone creates later is not covered by those grants.
const PgHbaSteampipeRules = `local all root      trust
host  all steampipe samehost trust
host  all steampipe all      scram-sha-256
`

// PgHbaRootRule is the rule that a pg_hba.conf the user manages must have, so that Steampipe
// can connect as root through the unix socket.
const PgHbaRootRule = "local all root trust"
