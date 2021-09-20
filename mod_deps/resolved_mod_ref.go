package mod_deps

import (
	"github.com/go-git/go-git/v5/plumbing"
	goVersion "github.com/hashicorp/go-version"
)

// ResolvedModRef is a struct to represent a resolved mod reference
type ResolvedModRef struct {
	// the FQN of the mod - also the Git URL of the mod repo
	Name string
	// the Git branch/tag
	GitReference plumbing.ReferenceName
	// the monotonic version - may be unknown for local or branch
	// although version will be monotonic, we can still use semver
	Version *goVersion.Version
	// the file path for local mods
	FilePath string
	// raw reference
	raw string
}

func NewResolvedModRef(modRef *ModRef, versionString string) (*ResolvedModRef, error) {
	res := &ResolvedModRef{
		Name: modRef.Name,
		raw:  modRef.raw,
		// these may be empty strings
		FilePath:     modRef.filePath,
		GitReference: plumbing.NewBranchReferenceName(modRef.branch),
	}

	// if we have a version, set the Version and GitTag properties from it
	if versionString != "" {
		v, err := goVersion.NewVersion(versionString)
		if err != nil {
			return nil, err
		}
		res.Version = v
		res.GitReference = plumbing.NewTagReferenceName(versionString)
	}
	return res, nil
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
