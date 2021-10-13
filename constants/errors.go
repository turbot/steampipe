package constants

import "strings"

// EEXISTS :: universal error string to denote that a resource already exists
const EEXISTS = "EEXISTS"

// ENOTEXISTS :: universal error string to denote that a resource does not exists
const ENOTEXISTS = "ENOTEXISTS"

func IsGRPCConnectivityError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "error reading from server: EOF")
}
