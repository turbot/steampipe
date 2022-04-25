package versionfile

type InstalledVersion struct {
	Name               string `json:"name"`
	Version            string `json:"version"`
	ImageDigest        string `json:"image_digest,omitempty"`
	BinaryDigest       string `json:"binary_digest,omitempty"`
	BinaryArchitecture string `json:"binary_arch,omitempty"`
	InstalledFrom      string `json:"installed_from,omitempty"`
	LastCheckedDate    string `json:"last_checked_date,omitempty"`
	InstallDate        string `json:"install_date,omitempty"`

	// legacy properties
	LegacyImageDigest     string `json:"imageDigest,omitempty"`
	LegacyInstalledFrom   string `json:"installedFrom,omitempty"`
	LegacyLastCheckedDate string `json:"lastCheckedDate,omitempty"`
	LegacyInstallDate     string `json:"installDate,omitempty"`
}

// MigrateLegacy migrates the legacy properties into new properties
func (f *InstalledVersion) MigrateLegacy() {
	f.ImageDigest = f.LegacyImageDigest
	f.InstalledFrom = f.LegacyInstalledFrom
	f.InstallDate = f.LegacyInstallDate
	f.LastCheckedDate = f.LegacyLastCheckedDate
}
