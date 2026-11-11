package constants

// PgHbaHeader starts every pg_hba.conf that Steampipe writes. Its first line is how
// Steampipe recognises a file that it wrote itself.
const PgHbaHeader = `# Written by Steampipe each time the database starts. DO NOT EDIT.
# Put your own rules in pg_hba.conf.d/<name>.conf. They come first, so they win.
`

// PgHbaSteampipeRules are Steampipe's own rules. They are written after every rule
// in pg_hba.conf.d.
//
// The database column is "all": at install, CONNECT is revoked from PUBLIC on the postgres
// and template1 databases, so the steampipe user can reach only its own database. A database
// that someone creates later is not covered by those grants.
const PgHbaSteampipeRules = `host all root      samehost trust
host all steampipe samehost trust
host all steampipe all      scram-sha-256
`

// PgHbaPreservedFileName is where a copy of a pg_hba.conf that someone else changed is
// kept, inside pg_hba.conf.d. A later copy replaces it.
const PgHbaPreservedFileName = "00-preserved.conf"
