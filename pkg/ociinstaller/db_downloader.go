package ociinstaller

import (
	"fmt"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/turbot/pipe-fittings/v2/ociinstaller"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

type dbDownloader struct {
	ociinstaller.OciDownloader[*dbImage, *dbImageConfig]
}

func (p *dbDownloader) EmptyConfig() *dbImageConfig {
	return &dbImageConfig{}
}

func newDbDownloader() *dbDownloader {
	res := &dbDownloader{}

	// create the base downloader, passing res as the image provider
	ociDownloader := ociinstaller.NewOciDownloader[*dbImage, *dbImageConfig](constants.BaseImageRef, SteampipeMediaTypeProvider{}, res)

	res.OciDownloader = *ociDownloader

	return res
}

func (p *dbDownloader) GetImageData(layers []ocispec.Descriptor) (*dbImage, error) {
	res := &dbImage{}

	// get the binary jar file
	mediaType, err := p.MediaTypesProvider.MediaTypeForPlatform("db")
	if err != nil {
		return nil, err
	}
	foundLayers := ociinstaller.FindLayersForMediaType(layers, mediaType[0])
	if len(foundLayers) != 1 {
		return nil, fmt.Errorf("invalid Image - should contain 1 installation file per platform, found %d", len(foundLayers))
	}
	res.ArchiveDir = foundLayers[0].Annotations["org.opencontainers.image.title"]

	// get the readme file info
	foundLayers = ociinstaller.FindLayersForMediaType(layers, MediaTypeDbDocLayer)
	if len(foundLayers) > 0 {
		res.ReadmeFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	// get the license file info
	foundLayers = ociinstaller.FindLayersForMediaType(layers, MediaTypeDbLicenseLayer)
	if len(foundLayers) > 0 {
		res.LicenseFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}
	return res, nil
}
