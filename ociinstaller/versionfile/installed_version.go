package versionfile

// LegacyInstalledVersion is the legacy db installed version info struct
type LegacyInstalledVersion struct {
	Name            string `json:"name"`
	Version         string `json:"version"`
	ImageDigest     string `json:"imageDigest"`
	InstalledFrom   string `json:"installedFrom"`
	LastCheckedDate string `json:"lastCheckedDate"`
	InstallDate     string `json:"installDate"`
}

type InstalledVersion struct {
	Name               string `json:"name"`
	Version            string `json:"version"`
	ImageDigest        string `json:"image_digest"`
	BinaryDigest       string `json:"binary_digest"`
	BinaryArchitecture string `json:"binary_arch"`
	InstalledFrom      string `json:"installed_from"`
	LastCheckedDate    string `json:"last_checked_date"`
	InstallDate        string `json:"install_date"`
}
