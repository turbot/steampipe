package mod_installer

import (
	"github.com/Masterminds/semver"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
	"github.com/turbot/steampipe/version"
)

type InstallationData struct {
	Name string
	// what version has been installed?
	Version *semver.Version
	// list of the mod versions available in Git (reverse sorted)
	AvailableVersions []*semver.Version
	//  all required constraints
	Constraints version.Constraints
	// TODO implement
	FilePath string
	Branch   string
}

func NewModInstallationData(modVersion *modconfig.ModVersion) (*InstallationData, error) {
	installationData := &InstallationData{
		Name:        modVersion.Name,
		Constraints: version.Constraints{},
	}
	sortedVersions, err := getTagVersionsFromGit(getGitUrl(modVersion.Name))
	if err != nil {
		return nil, err
	}
	// we have not cached this - fetch from git
	installationData.AvailableVersions = sortedVersions
	installationData.addConstraint(modVersion)
	return installationData, nil
}

// add the constraint from this mod version to our constraints
func (m *InstallationData) addConstraint(modVersion *modconfig.ModVersion) {
	m.Constraints.Add(modVersion.VersionConstraint)
}
