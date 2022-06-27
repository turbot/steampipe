//go:build linux && amd64
// +build linux,amd64

package platform

var Paths = PlatformPaths{
	TarFileName:         "postgres-linux-x86_64.txz",
	InitDbExecutable:    "initdb",
	PostgresExecutable:  "postgres",
	PgDumpExecutable:    "pg_dump",
	PgRestoreExecutable: "pg_restore",
}
