//go:build linux && arm64
// +build linux,arm64

package platform

var Paths = PlatformPaths{
	TarFileName:         "postgres-linux-arm_64.txz",
	InitDbExecutable:    "initdb",
	PostgresExecutable:  "postgres",
	PgDumpExecutable:    "pg_dump",
	PgRestoreExecutable: "pg_restore",
}
