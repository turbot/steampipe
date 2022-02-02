// Package version :: The version package provides a location to set the release versions for all
// packages to consume, without creating import cycles.
//
// This package should not import any other steampipe packages.
//
package version

import (
	"fmt"

	"github.com/Masterminds/semver"
)

/**
We should fill in the `steampipeVersion` and `prerelease` variables using ldflags during build

See https://blog.alexellis.io/inject-build-time-vars-golang/
Also https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications
**/

// The main version number that is being run at the moment.
var steampipeVersion = "0.13.0"

// A pre-release marker for the version. If this is "" (empty string)
// then it means that it is a final release. Otherwise, this is a pre-release
// such as "dev" (in development), "beta", "rc1", etc.
var prerelease = "alpha.3"

// SteampipeVersion is an instance of semver.Version. This has the secondary
// benefit of verifying during tests and init time that our version is a
// proper semantic version, which should always be the case.
var SteampipeVersion *semver.Version

func init() {
	versionString := steampipeVersion
	if prerelease != "" {
		versionString = fmt.Sprintf("%s-%s", steampipeVersion, prerelease)
	}
	SteampipeVersion = semver.MustParse(versionString)
}
