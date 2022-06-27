package ociinstaller

import (
	"fmt"
	"runtime"

	"github.com/turbot/steampipe/pkg/constants"
	"github.com/turbot/steampipe/pkg/utils"
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

	MediaTypeAssetReportLayer = "application/vnd.turbot.steampipe.assets.report.layer.v1+tar"
)

// MediaTypeForPlatform returns media types for binaries for this OS and architecture
// and it's fallbacks in order of priority
func MediaTypeForPlatform(imageType ImageType) ([]string, error) {
	layerFmtGzip := "application/vnd.turbot.steampipe.%s.%s-%s.layer.v1+gzip"
	layerFmtTar := "application/vnd.turbot.steampipe.%s.%s-%s.layer.v1+tar"

	arch := runtime.GOARCH
	switch imageType {
	case ImageTypeDatabase:
		return []string{fmt.Sprintf(layerFmtTar, imageType, runtime.GOOS, arch)}, nil
	case ImageTypeFdw:
		// detect the underlying architecture(amd64/arm64)
		// we have to do this rather than just using runtime.GOARCH, because runtime.GOARCH does not give us
		// the actual underlying architecture of the system(GOARCH can be changed during runtime)
		arch, err := utils.UnderlyingArch()
		if err != nil {
			return nil, err
		}
		return []string{fmt.Sprintf(layerFmtGzip, imageType, runtime.GOOS, arch)}, nil
	case ImageTypePlugin:
		pluginMediaTypes := []string{fmt.Sprintf(layerFmtGzip, imageType, runtime.GOOS, arch)}
		if runtime.GOOS == constants.OSDarwin && arch == constants.ArchARM64 {
			// add the amd64 layer as well, so that we can fall back to it
			// this is required for plugins which don't have an arm64 build yet
			pluginMediaTypes = append(pluginMediaTypes, fmt.Sprintf(layerFmtGzip, imageType, runtime.GOOS, constants.ArchAMD64))
		}
		return pluginMediaTypes, nil
	}
	// there are cases(dashboard commands) where we have a different imageType, we need to return empty
	// in such cases and not return error
	return []string{}, nil
}

// SharedMediaTypes returns media types that are NOT specific to the os and arch (readmes, control files, etc)
func SharedMediaTypes(imageType ImageType) []string {
	switch imageType {
	case ImageTypeAssets:
		return []string{MediaTypeAssetReportLayer}
	case ImageTypeDatabase:
		return []string{MediaTypeDbDocLayer, MediaTypeDbLicenseLayer}
	case ImageTypeFdw:
		return []string{MediaTypeFdwDocLayer, MediaTypeFdwLicenseLayer, MediaTypeFdwControlLayer, MediaTypeFdwSqlLayer}
	case ImageTypePlugin:
		return []string{MediaTypePluginDocsLayer, MediaTypePluginSpcLayer, MediaTypePluginLicenseLayer}
	}
	return nil
}

// ConfigMediaTypes :: returns media types for OCI $config data ( in the config, not a layer)
func ConfigMediaTypes() []string {
	return []string{MediaTypeConfig, MediaTypePluginConfig}
}
