package parse

import (
	goVersion "github.com/hashicorp/go-version"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

type InstalledMod struct {
	Mod     *modconfig.Mod
	Version *goVersion.Version
}
