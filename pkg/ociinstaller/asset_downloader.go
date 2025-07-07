package ociinstaller

import (
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/turbot/pipe-fittings/v2/ociinstaller"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

type assetsDownloader struct {
	ociinstaller.OciDownloader[*assetsImage, *assetsImageConfig]
}

func (p *assetsDownloader) EmptyConfig() *assetsImageConfig {
	return &assetsImageConfig{}
}

func newAssetDownloader() *assetsDownloader {
	res := &assetsDownloader{}

	// create the base downloader, passing res as the image provider
	ociDownloader := ociinstaller.NewOciDownloader[*assetsImage, *assetsImageConfig](constants.BaseImageRef, SteampipeMediaTypeProvider{}, res)

	res.OciDownloader = *ociDownloader

	return res
}

func (p *assetsDownloader) GetImageData(layers []ocispec.Descriptor) (*assetsImage, error) {
	var assetImage assetsImage

	// get the report dir
	foundLayers := ociinstaller.FindLayersForMediaType(layers, MediaTypeAssetReportLayer)
	if len(foundLayers) > 0 {
		assetImage.ReportUI = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	return &assetImage, nil
}
