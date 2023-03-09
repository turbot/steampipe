//go:build linux && arm
// +build linux,arm

package platform

var Paths = PlatformPaths{
	TarFileName:         "postgres-linux-arm_32.txz",
	InitDbExecutable:    "initdb",
	PostgresExecutable:  "postgres",
	PgDumpExecutable:    "pg_dump",
	PgRestoreExecutable: "pg_restore",
}
