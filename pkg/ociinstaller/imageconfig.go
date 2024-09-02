package ociinstaller

import (
	"encoding/json"
)

const DefaultConfigSchema string = "2020-11-18"

type OciImageConfig interface {
	Name()
}

type OciConfig interface {
	GetSchemaVersion() string
	SetSchemaVersion(string)
}

type OciConfigBase struct {
	SchemaVersion string `json:"schemaVersion"`
}

func (c *OciConfigBase) GetSchemaVersion() string {
	return c.SchemaVersion
}
func (c *OciConfigBase) SetSchemaVersion(version string) {
	c.SchemaVersion = version
}

type configPlugin struct {
	OciConfigBase
	Plugin struct {
		Name         string `json:"name,omitempty"`
		Organization string `json:"organization,omitempty"`
		Version      string `json:"version"`
	}
}

type configDb struct {
	OciConfigBase
	Database struct {
		Name         string `json:"name,omitempty"`
		Organization string `json:"organization,omitempty"`
		Version      string `json:"version"`
		DBVersion    string `json:"dbVersion,omitempty"`
	}
}

type configFdw struct {
	OciConfigBase
	Fdw struct {
		Name         string `json:"name,omitempty"`
		Organization string `json:"organization,omitempty"`
		Version      string `json:"version"`
	}
}

func newSteampipeImageConfig(configBytes []byte, imageType ImageType) (OciConfig, error) {
	var target OciConfig
	switch imageType {
	case ImageTypeDatabase:
		target = &configDb{}
	case ImageTypeFdw:
		target = &configFdw{}
	case ImageTypePlugin:
		target = &configPlugin{}
	}

	if err := json.Unmarshal(configBytes, target); err != nil {
		return nil, err
	}

	if target.GetSchemaVersion() == "" {
		target.SetSchemaVersion(DefaultConfigSchema)
	}
	return target, nil
}
