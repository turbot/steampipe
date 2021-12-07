package mod_installer

import (
	"github.com/turbot/steampipe/steampipeconfig/modconfig"

	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v5/plumbing"
)

// ResolvedModRef is a struct to represent a resolved mod git reference
type ResolvedModRef struct {
	// the FQN of the mod - also the Git URL of the mod repo
	Name string
	// the mod version
	Version *semver.Version

	// the Git branch/tag
	GitReference plumbing.ReferenceName
	// the file path for local mods
	FilePath string
}

func NewResolvedModRef(requiredModVersion *modconfig.ModVersionConstraint, version *semver.Version) (*ResolvedModRef, error) {
	res := &ResolvedModRef{
		Name:    requiredModVersion.Name,
		Version: version,
		// this may be empty strings
		FilePath: requiredModVersion.FilePath,
	}
	if res.FilePath == "" {
		res.setGitReference()
	}

	return res, nil
}

func (r *ResolvedModRef) setGitReference() {
	// TODO handle branches
	//if modVersion.Branch != "" {
	//	r.GitReference = plumbing.NewBranchReferenceName(modVersion.Branch)
	//	// NOTE: we need to set version from branch
	//	return
	//}

	// NOTE: use the original version string - this will be the tag name
	r.GitReference = plumbing.NewTagReferenceName(r.Version.Original())
}

// FullName returns name in the format <dependency name>@v<dependencyVersion>
func (r *ResolvedModRef) FullName() string {
	return modVersionFullName(r.Name, r.Version)
}
