// +build windows
// +build amd64

package platform

var Paths = PlatformPaths{
	TarFileName:        "postgres-windows-x86_64.txz",
	InitDbExecutable:   "initdb.exe",
	PostgresExecutable: "postgres.exe",
}
