package mod_installer

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v5/plumbing"
)

// ResolvedModRef is a struct to represent a resolved mod git reference
type ResolvedModRef struct {
	// the FQN of the mod - also the Git URL of the mod repo
	Name string
	// the Git branch/tag
	GitReference plumbing.ReferenceName
	// the mod version
	Version *semver.Version
	// the file path for local mods
	FilePath string
}

func NewResolvedModRef(installationData *InstallationData, version *semver.Version) (*ResolvedModRef, error) {
	res := &ResolvedModRef{
		Name: installationData.Name,

		// this may be empty strings
		FilePath: installationData.FilePath,
	}
	if res.FilePath == "" {
		res.SetGitReference(version)
	}

	return res, nil
}

func (r *ResolvedModRef) SetGitReference(version *semver.Version) {
	r.Version = version
	// TODO handle branches
	//if modVersion.Branch != "" {
	//	r.GitReference = plumbing.NewBranchReferenceName(modVersion.Branch)
	//	// NOTE: we need to set version from branch
	//	return
	//}

	// NOTE: use the original version string - this will be the tag name
	r.GitReference = plumbing.NewTagReferenceName(version.Original())
}

// FullName returns name in the format <dependency name>@v<dependencyVersion>
func (r *ResolvedModRef) FullName() string {
	return fmt.Sprintf("%s@%s", r.Name, r.Version.Original())
}
