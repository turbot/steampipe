//go:build darwin && arm64
// +build darwin,arm64

package platform

var Paths = PlatformPaths{
	TarFileName:        "postgres-darwin-arm_64.txz",
	InitDbExecutable:   "initdb",
	PostgresExecutable: "postgres",
}
