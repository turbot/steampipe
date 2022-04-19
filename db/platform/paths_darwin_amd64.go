//go:build darwin && amd64
// +build darwin,amd64

package platform

var Paths = PlatformPaths{
	TarFileName:         "postgres-darwin-x86_64.txz",
	InitDbExecutable:    "initdb",
	PostgresExecutable:  "postgres",
	PgDumpExecutable:    "pg_dump",
	PgRestoreExecutable: "pg_restore",
}
