package ociinstaller

import (
	"fmt"
	"runtime"

	"github.com/turbot/pipe-fittings/v2/constants"
	"github.com/turbot/pipe-fittings/v2/ociinstaller"
	"github.com/turbot/pipe-fittings/v2/utils"
)

// Steampipe Media Types
const (
	MediaTypeDbDocLayer       = "application/vnd.turbot.steampipe.db.doc.layer.v1+text"
	MediaTypeDbLicenseLayer   = "application/vnd.turbot.steampipe.db.license.layer.v1+text"
	MediaTypeFdwDocLayer      = "application/vnd.turbot.steampipe.fdw.doc.layer.v1+text"
	MediaTypeFdwLicenseLayer  = "application/vnd.turbot.steampipe.fdw.license.layer.v1+text"
	MediaTypeFdwControlLayer  = "application/vnd.turbot.steampipe.fdw.control.layer.v1+text"
	MediaTypeFdwSqlLayer      = "application/vnd.turbot.steampipe.fdw.sql.layer.v1+text"
	MediaTypeAssetReportLayer = "application/vnd.turbot.steampipe.assets.report.layer.v1+tar"
)

type SteampipeMediaTypeProvider struct{}

func (p SteampipeMediaTypeProvider) GetAllMediaTypes(imageType ociinstaller.ImageType) ([]string, error) {
	m, err := p.MediaTypeForPlatform(imageType)
	if err != nil {
		return nil, err
	}
	s := p.SharedMediaTypes(imageType)
	c := p.ConfigMediaTypes()
	return append(append(m, s...), c...), nil
}

// MediaTypeForPlatform returns media types for binaries for this OS and architecture
// and it's fallbacks in order of priority
func (SteampipeMediaTypeProvider) MediaTypeForPlatform(imageType ociinstaller.ImageType) ([]string, error) {
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
	case ociinstaller.ImageTypePlugin:
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
func (SteampipeMediaTypeProvider) SharedMediaTypes(imageType ociinstaller.ImageType) []string {
	switch imageType {
	case ImageTypeAssets:
		return []string{MediaTypeAssetReportLayer}
	case ImageTypeDatabase:
		return []string{MediaTypeDbDocLayer, MediaTypeDbLicenseLayer}
	case ImageTypeFdw:
		return []string{MediaTypeFdwDocLayer, MediaTypeFdwLicenseLayer, MediaTypeFdwControlLayer, MediaTypeFdwSqlLayer}
	case ociinstaller.ImageTypePlugin:
		return []string{ociinstaller.MediaTypePluginSpcLayer(), ociinstaller.MediaTypePluginLicenseLayer()}
	}
	return nil
}

// ConfigMediaTypes :: returns media types for OCI $config data ( in the config, not a layer)
func (SteampipeMediaTypeProvider) ConfigMediaTypes() []string {
	return []string{ociinstaller.MediaTypeConfig(), ociinstaller.MediaTypePluginConfig()}
}
