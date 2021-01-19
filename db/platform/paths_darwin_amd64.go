// +build darwin
// +build amd64

package platform

var Paths = PlatformPaths{
	TarFileName:        "postgres-darwin-x86_64.txz",
	InitDbExecutable:   "initdb",
	PostgresExecutable: "postgres",
}
