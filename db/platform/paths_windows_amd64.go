//go:build windows && amd64
// +build windows,amd64

package platform

var Paths = PlatformPaths{
	TarFileName:         "postgres-windows-x86_64.txz",
	InitDbExecutable:    "initdb.exe",
	PostgresExecutable:  "postgres.exe",
	PgDumpExecutable:    "pg_dump",
	PgRestoreExecutable: "pg_restore",
}
