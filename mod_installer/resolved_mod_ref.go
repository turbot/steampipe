package mod_installer

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	goVersion "github.com/hashicorp/go-version"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
)

// ResolvedModRef is a struct to represent a resolved mod git reference
type ResolvedModRef struct {
	// the FQN of the mod - also the Git URL of the mod repo
	Name string
	// the Git branch/tag
	GitReference plumbing.ReferenceName
	// the mod version
	Version *goVersion.Version
	// the file path for local mods
	FilePath string
}

func NewResolvedModRef(modVersion *modconfig.ModVersion, version *goVersion.Version) (*ResolvedModRef, error) {
	res := &ResolvedModRef{
		Name: modVersion.Name,

		// this may be empty strings
		FilePath: modVersion.FilePath,
	}
	if res.FilePath == "" {
		res.SetGitReference(modVersion, version)
	}

	return res, nil
}

func (r *ResolvedModRef) SetGitReference(modVersion *modconfig.ModVersion, version *goVersion.Version) {
	r.Version = version
	if modVersion.Branch != "" {
		r.GitReference = plumbing.NewBranchReferenceName(modVersion.Branch)
		// NOTE: we need to set version from branch
		return
	}

	// NOTE: use the original version string - this will be the tag name
	r.GitReference = plumbing.NewTagReferenceName(version.Original())
}

// FullName returns name in the format <dependency name>@v<dependencyVersion>
func (r *ResolvedModRef) FullName() string {
	segments := r.Version.Segments()
	return fmt.Sprintf("%s@v%d.%d", r.Name, segments[0], segments[1])
}

// SatisfiesVersionConstraint return whether this resolved ref satisfies a version constraint
func (r *ResolvedModRef) SatisfiesVersionConstraint(versionConstraint *goVersion.Version) bool {
	// if we do not have a version set, then we cannot satisfy a version constraint
	// this may happen if we are a local file or unversioned branch
	if r.Version == nil {
		return false
	}

	return r.Version.GreaterThanOrEqual(versionConstraint)
}
