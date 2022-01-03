package versionmap

import (
	"github.com/Masterminds/semver"
)

// VersionMap represents a map of semver versions, keyed by dependency name
type VersionMap map[string]*semver.Version
