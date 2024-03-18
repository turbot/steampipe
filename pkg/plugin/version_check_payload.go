package plugin

import "fmt"

type versionCheckPayload interface {
	getMapKey() string
}

// the payload that travels to-and-fro between steampipe and the server
type versionCheckCorePayload struct {
	Org        string `json:"org"`
	Name       string `json:"name"`
	Constraint string `json:"constraint"`
	Version    string `json:"version"`
}

func (v *versionCheckCorePayload) getMapKey() string {
	return fmt.Sprintf("%s/%s/%s", v.Org, v.Name, v.Constraint)
}
