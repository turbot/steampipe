// +build linux
// +build amd64

package platform

var Paths = PlatformPaths{
	TarFileName:        "postgres-linux-x86_64.txz",
	InitDbExecutable:   "initdb",
	PostgresExecutable: "postgres",
}
