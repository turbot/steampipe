// +build linux
// +build arm

package platform

var Paths = PlatformPaths{
	TarFileName:        "postgres-linux-arm_32.txz",
	InitDbExecutable:   "initdb",
	PostgresExecutable: "postgres",
}
