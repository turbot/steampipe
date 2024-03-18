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

type responseManifestAnnotations map[string]string
type responseManifestConfig struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int    `json:"size"`
}
type responseManifestLayer struct {
	responseManifestConfig
	Annotations responseManifestAnnotations `json:"annotations"`
}
type responseManifest struct {
	SchemaVersion int                         `json:"schemaVersion"`
	Config        responseManifestConfig      `json:"config"`
	Layers        []responseManifestLayer     `json:"layers"`
	Annotations   responseManifestAnnotations `json:"annotations"`
}
type versionCheckResponsePayload struct {
	versionCheckCorePayload
	Digest   string           `json:"digest"`
	Manifest responseManifest `json:"manifest"`
}
