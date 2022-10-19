package constants

import (
	"errors"
)

var MissingCloudTokenError = errors.New("to share snapshots to Steampipe Cloud, cloud token must be set")

// EEXISTS is the universal error string to denote that a resource already exists
const EEXISTS = "EEXISTS"

// ENOTEXISTS is the universal error string to denote that a resource does not exists
const ENOTEXISTS = "ENOTEXISTS"
