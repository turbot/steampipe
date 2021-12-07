package mod_installer

import (
	"fmt"

	"github.com/Masterminds/semver"
)

func modVersionFullName(name string, version *semver.Version) string {
	return fmt.Sprintf("%s@%s", name, version.Original())
}
