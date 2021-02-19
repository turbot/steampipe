// The version package provides a location to set the release versions for all
// packages to consume, without creating import cycles.
//
// This package should not import any other steampipe packages.
package version

import (
	"fmt"

	goVersion "github.com/hashicorp/go-version"
)

// The main version number that is being run at the moment.
var steampipeVersion = "0.2.0"

// A pre-release marker for the version. If this is "" (empty string)
// then it means that it is a final release. Otherwise, this is a pre-release
// such as "dev" (in development), "beta", "rc1", etc.
var prerelease = ""

// The git commit that was compiled. This will be filled in by the compiler.
// See https://blog.alexellis.io/inject-build-time-vars-golang/
// Also https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications
var gitCommit = "gitcommit"

// semVer is an instance of version.Version. This has the secondary
// benefit of verifying during tests and init time that our version is a
// proper semantic version, which should always be the case.
var semVer *goVersion.Version

func init() {
	semVer = goVersion.Must(goVersion.NewVersion(steampipeVersion))
}

// header is the header name used to send the current steampipe version
// in http requests.
const header = "Steampipe-Version"

// String returns the complete version string, including prerelease
func String() string {
	versionString := steampipeVersion
	if prerelease != "" {
		versionString = fmt.Sprintf("%s-%s", versionString, prerelease)
		if gitCommit != "" {
			// we should not include git commit if this is not prerelease
			versionString = fmt.Sprintf("%s+%s", versionString, gitCommit)
		}
	}
	return versionString
}
