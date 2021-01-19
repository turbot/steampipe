// +build linux
// +build 386

package platform

var Paths = PlatformPaths{
	TarFileName:        "postgres-linux-x86_32.txz",
	InitDbExecutable:   "initdb",
	PostgresExecutable: "postgres",
}
