package mod_installer

import goVersion "github.com/hashicorp/go-version"

type ModInstallationData struct {
	Name string
	// reverse order versions available
	Versions []*goVersion.Version
}
