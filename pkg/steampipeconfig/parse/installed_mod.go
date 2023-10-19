package parse

import (
	"github.com/Masterminds/semver/v3"
	"github.com/turbot/pipe-fittings/modconfig"
)

type InstalledMod struct {
	Mod     *modconfig.Mod
	Version *semver.Version
}
