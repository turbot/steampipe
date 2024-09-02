package ociinstaller

import "github.com/turbot/pipe-fittings/ociinstaller"

type ConfigDb struct {
	ociinstaller.OciConfigBase
	Database struct {
		Name         string `json:"name,omitempty"`
		Organization string `json:"organization,omitempty"`
		Version      string `json:"version"`
		DBVersion    string `json:"dbVersion,omitempty"`
	}
}

type ConfigFdw struct {
	ociinstaller.OciConfigBase
	Fdw struct {
		Name         string `json:"name,omitempty"`
		Organization string `json:"organization,omitempty"`
		Version      string `json:"version"`
	}
}
