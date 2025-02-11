package ociinstaller

import "github.com/turbot/pipe-fittings/v2/ociinstaller"

type dbImage struct {
	ArchiveDir  string
	ReadmeFile  string
	LicenseFile string
}

func (s *dbImage) Type() ociinstaller.ImageType {
	return ImageTypeDatabase
}

type dbImageConfig struct {
	ociinstaller.OciConfigBase
	Database struct {
		Name         string `json:"name,omitempty"`
		Organization string `json:"organization,omitempty"`
		Version      string `json:"version"`
		DBVersion    string `json:"dbVersion,omitempty"`
	}
}
