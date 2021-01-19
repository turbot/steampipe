// +build linux
// +build arm64

package platform

var Paths = PlatformPaths{
	TarFileName:        "postgres-linux-arm_64.txz",
	InitDbExecutable:   "initdb",
	PostgresExecutable: "postgres",
}
