package mod_installer

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
)

// ModRef is a struct to represent an unresolved mod reference
type ModRef struct {
	// the Git URL of the mod repo
	Name string
	// the version constraint of the mod
	versionConstraint *semver.Version
	// the branch to use
	branch string
	// the local file location to use
	filePath string
	// raw reference
	raw string
}

func NewModRef(modRef string) (*ModRef, error) {
	split := strings.Split(modRef, "@")
	if len(split) > 2 {
		return nil, fmt.Errorf("invalid mod ref %s", modRef)
	}
	res := &ModRef{
		raw:  modRef,
		Name: split[0],
	}
	if len(split) == 2 {
		res.setVersion(split[1])
	}

	return res, nil
}

func (r *ModRef) setVersion(versionString string) {
	if strings.HasPrefix(versionString, "file:") {
		r.filePath = versionString
		return
	}
	// does the verison parse as a semver version
	if v, err := semver.NewVersion(versionString); err == nil {
		r.versionConstraint = v
		return
	}

	// otherwise assume it is a branch
	r.branch = versionString
}
