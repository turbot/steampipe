package plugin

import "fmt"

type versionCheckPayload interface {
	getMapKey() string
}

// the payload that travels to-and-fro between steampipe and the server
type versionCheckRequestPayload struct {
	Org     string `json:"org"`
	Name    string `json:"name"`
	Stream  string `json:"stream"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

func (v *versionCheckRequestPayload) getMapKey() string {
	return fmt.Sprintf("%s/%s/%s", v.Org, v.Name, v.Stream)
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
	versionCheckRequestPayload
	Manifest responseManifest `json:"manifest"`
}
