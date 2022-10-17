package steampipeconfig

import "strings"

// IsCloudWorkspaceIdentifier returns whether name is a cloud workspace identifier
// of the form: {identity_handle}/{workspace_handle},
func IsCloudWorkspaceIdentifier(name string) bool {
	return len(strings.Split(name, "/")) == 2
}
