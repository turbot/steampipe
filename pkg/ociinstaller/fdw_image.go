package ociinstaller

import "github.com/turbot/pipe-fittings/v2/ociinstaller"

type fdwImage struct {
	BinaryFile  string
	ReadmeFile  string
	LicenseFile string
	ControlFile string
	SqlFile     string
}

func (s *fdwImage) Type() ociinstaller.ImageType {
	return ImageTypeFdw
}

type FdwImageConfig struct {
	ociinstaller.OciConfigBase
	Fdw struct {
		Name         string `json:"name,omitempty"`
		Organization string `json:"organization,omitempty"`
		Version      string `json:"version"`
	}
}
