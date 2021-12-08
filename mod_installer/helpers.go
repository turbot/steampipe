package mod_installer

import (
	"fmt"

	"github.com/turbot/steampipe/version"

	"github.com/Masterminds/semver"
)

func modVersionFullName(name string, version *semver.Version) string {
	return fmt.Sprintf("%s@%s", name, version.Original())
}

func getVersionSatisfyingConstraint(constraint *version.Constraints, availableVersions []*semver.Version) *semver.Version {
	// search the reverse sorted versions, finding the highest version which satisfies ALL constraints
	for _, version := range availableVersions {
		if constraint.Check(version) {
			return version
		}
	}
	return nil
}
