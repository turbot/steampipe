package ociinstaller

import (
	"encoding/json"
)

const DefaultConfigSchema string = "2020-11-18"

type config struct {
	SchemaVersion string        `json:"schemaVersion"`
	Plugin        *configPlugin `json:"plugin,omitempty"`
	Database      *configDb     `json:"db,omitempty"`
	Fdw           *configFdw    `json:"fdw,omitempty"`
}

type configPlugin struct {
	Name         string `json:"name,omitempty"`
	Organization string `json:"organization,omitempty"`
	Version      string `json:"version"`
}

type configDb struct {
	Name         string `json:"name,omitempty"`
	Organization string `json:"organization,omitempty"`
	Version      string `json:"version"`
	DBVersion    string `json:"dbVersion,omitempty"`
}

type configFdw struct {
	Name         string `json:"name,omitempty"`
	Organization string `json:"organization,omitempty"`
	Version      string `json:"version"`
}

func newSteampipeImageConfig(configBytes []byte) (*config, error) {
	configData := &config{
		Plugin:   &configPlugin{},
		Database: &configDb{},
		Fdw:      &configFdw{},
	}
	if err := json.Unmarshal(configBytes, configData); err != nil {
		return nil, err
	}

	if configData.SchemaVersion == "" {
		configData.SchemaVersion = DefaultConfigSchema
	}
	return configData, nil
}
