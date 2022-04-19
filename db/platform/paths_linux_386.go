//go:build linux && 386
// +build linux,386

package platform

var Paths = PlatformPaths{
	TarFileName:         "postgres-linux-x86_32.txz",
	InitDbExecutable:    "initdb",
	PostgresExecutable:  "postgres",
	PgDumpExecutable:    "pg_dump",
	PgRestoreExecutable: "pg_restore",
}
