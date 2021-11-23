package ociinstaller

import (
	"fmt"
	"runtime"
)

// Steampipe Media Types
const (
	MediaTypeConfig = "application/vnd.turbot.steampipe.config.v1+json"

	//deprecate this....
	MediaTypePluginConfig = "application/vnd.turbot.steampipe.plugin.config.v1+json"

	MediaTypePluginDarwinAmd64Layer  = "application/vnd.turbot.steampipe.plugin.darwin-amd64.layer.v1+gzip"
	MediaTypePluginLinuxAmd64Layer   = "application/vnd.turbot.steampipe.plugin.linux-amd64.layer.v1+gzip"
	MediaTypePluginWindowsAmd64Layer = "application/vnd.turbot.steampipe.plugin.windows-amd64.layer.v1+gzip"
	MediaTypePluginDarwinArm64Layer  = "application/vnd.turbot.steampipe.plugin.darwin-arm64.layer.v1+gzip"
	MediaTypePluginLinuxArm64Layer   = "application/vnd.turbot.steampipe.plugin.linux-arm64.layer.v1+gzip"
	MediaTypePluginWindowsArm64Layer = "application/vnd.turbot.steampipe.plugin.windows-arm64.layer.v1+gzip"
	MediaTypePluginLicenseLayer      = "application/vnd.turbot.steampipe.plugin.license.layer.v1+text"
	MediaTypePluginDocsLayer         = "application/vnd.turbot.steampipe.plugin.docs.layer.v1+tar"
	MediaTypePluginSpcLayer          = "application/vnd.turbot.steampipe.plugin.spc.layer.v1+tar"

	MediaTypeDbDarwinAmd64Layer  = "application/vnd.turbot.steampipe.db.darwin-amd64.layer.v1+tar"
	MediaTypeDbLinuxAmd64Layer   = "application/vnd.turbot.steampipe.db.linux-amd64.layer.v1+tar"
	MediaTypeDbWindowsAmd64Layer = "application/vnd.turbot.steampipe.db.windows-amd64.layer.v1+tar"
	MediaTypeDbDarwinArm64Layer  = "application/vnd.turbot.steampipe.db.darwin-arm64.layer.v1+tar"
	MediaTypeDbLinuxArm64Layer   = "application/vnd.turbot.steampipe.db.linux-arm64.layer.v1+tar"
	MediaTypeDbWindowsArm64Layer = "application/vnd.turbot.steampipe.db.windows-arm64.layer.v1+tar"
	MediaTypeDbDocLayer          = "application/vnd.turbot.steampipe.db.doc.layer.v1+text"
	MediaTypeDbLicenseLayer      = "application/vnd.turbot.steampipe.db.license.layer.v1+text"

	MediaTypeFdwDarwinAmd64Layer  = "application/vnd.turbot.steampipe.fdw.darwin-amd64.layer.v1+gzip"
	MediaTypeFdwLinuxAmd64Layer   = "application/vnd.turbot.steampipe.fdw.linux-amd64.layer.v1+gzip"
	MediaTypeFdwWindowsAmd64Layer = "application/vnd.turbot.steampipe.fdw.windows-amd64.layer.v1+gzip"
	MediaTypeFdwDarwinArm64Layer  = "application/vnd.turbot.steampipe.fdw.darwin-arm64.layer.v1+gzip"
	MediaTypeFdwLinuxArm64Layer   = "application/vnd.turbot.steampipe.fdw.linux-arm64.layer.v1+gzip"
	MediaTypeFdwWindowsArm64Layer = "application/vnd.turbot.steampipe.fdw.windows-arm64.layer.v1+gzip"
	MediaTypeFdwDocLayer          = "application/vnd.turbot.steampipe.fdw.doc.layer.v1+text"
	MediaTypeFdwLicenseLayer      = "application/vnd.turbot.steampipe.fdw.license.layer.v1+text"

	MediaTypeFdwControlLayer = "application/vnd.turbot.steampipe.fdw.control.layer.v1+text"
	MediaTypeFdwSqlLayer     = "application/vnd.turbot.steampipe.fdw.sql.layer.v1+text"
)

// MediaTypeForPlatform returns media types for binaries for this OS and architecture
func MediaTypeForPlatform(imageType string) string {
	// we do not (yet) support Arm for the database, FDW or plugins - on M1 macs Rosetta will emulate this for us
	arch := "amd64"
	switch imageType {
	case "db":
		return fmt.Sprintf("application/vnd.turbot.steampipe.%s.%s-%s.layer.v1+tar", imageType, runtime.GOOS, arch)
	case "fdw":
		return fmt.Sprintf("application/vnd.turbot.steampipe.%s.%s-%s.layer.v1+gzip", imageType, runtime.GOOS, arch)
	case "plugin":
		return fmt.Sprintf("application/vnd.turbot.steampipe.%s.%s-%s.layer.v1+gzip", imageType, runtime.GOOS, arch)
	}
	return ""
}

// SharedMediaTypes returns media types that are NOT specific to the os and arch (readmes, control files, etc)
func SharedMediaTypes(imageType string) []string {
	switch imageType {
	case "db":
		return []string{MediaTypeDbDocLayer, MediaTypeDbLicenseLayer}
	case "fdw":
		return []string{MediaTypeFdwDocLayer, MediaTypeFdwLicenseLayer, MediaTypeFdwControlLayer, MediaTypeFdwSqlLayer}
	case "plugin":
		return []string{MediaTypePluginDocsLayer, MediaTypePluginSpcLayer, MediaTypePluginLicenseLayer}
	}
	return nil
}

// ConfigMediaTypes :: returns media types for OCI $config data ( in the config, not a layer)
func ConfigMediaTypes() []string {
	return []string{MediaTypeConfig, MediaTypePluginConfig}
}
