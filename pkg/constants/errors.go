package constants

import (
	"errors"
)

var MissingCloudTokenError = errors.New("no token found, please run 'steampipe login'")

// EEXISTS is the universal error string to denote that a resource already exists
const EEXISTS = "EEXISTS"

// ENOTEXISTS is the universal error string to denote that a resource does not exists
const ENOTEXISTS = "ENOTEXISTS"
