package mod_installer

import (
	"fmt"

	"github.com/go-git/go-git/v5/plumbing"
	goVersion "github.com/hashicorp/go-version"
	"github.com/turbot/steampipe/steampipeconfig/modconfig"
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
}

func NewResolvedModRef(modVersion *modconfig.ModVersion) (*ResolvedModRef, error) {
	res := &ResolvedModRef{
		Name: modVersion.Name,

		// this may be empty strings
		FilePath: modVersion.FilePath,
	}
	if res.FilePath == "" {
		// TODO we currently only support explicit (i.e. minor) versions
		// if the mod version has either a version constraint or branch, set the git ref
		res.SetGitReference(modVersion)
	}

	return res, nil
}

func (r *ResolvedModRef) SetGitReference(modVersion *modconfig.ModVersion) {

	if modVersion.Branch != "" {
		r.GitReference = plumbing.NewBranchReferenceName(modVersion.Branch)
		// TODO set version from branch
		return
	}

	// so there is aversion constraint
	// TODO if it is just a major constraint, we need to find the latest version in the major
	// for now assume it is a full version
	// NOTE: we cannot just ToString the version as we need the 'v' at the beginning
	r.GitReference = plumbing.NewTagReferenceName(modVersion.VersionString)
	r.Version = modVersion.VersionConstraint
}

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
