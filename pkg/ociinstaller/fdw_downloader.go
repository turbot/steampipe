package ociinstaller

import (
	"fmt"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/turbot/pipe-fittings/v2/ociinstaller"
	"github.com/turbot/steampipe/v2/pkg/constants"
)

type fdwDownloader struct {
	ociinstaller.OciDownloader[*fdwImage, *FdwImageConfig]
}

func (p *fdwDownloader) EmptyConfig() *FdwImageConfig {
	return &FdwImageConfig{}
}

func newFdwDownloader() *fdwDownloader {
	res := &fdwDownloader{}

	// create the base downloader, passing res as the image provider
	ociDownloader := ociinstaller.NewOciDownloader[*fdwImage, *FdwImageConfig](constants.BaseImageRef, SteampipeMediaTypeProvider{}, res)

	res.OciDownloader = *ociDownloader

	return res
}

func (p *fdwDownloader) GetImageData(layers []ocispec.Descriptor) (*fdwImage, error) {
	res := &fdwImage{}

	// get the binary (steampipe-postgres-fdw.so) info
	mediaType, err := p.MediaTypesProvider.MediaTypeForPlatform("fdw")
	if err != nil {
		return nil, err
	}
	foundLayers := ociinstaller.FindLayersForMediaType(layers, mediaType[0])
	if len(foundLayers) != 1 {
		return nil, fmt.Errorf("invalid image - image should contain 1 binary file per platform, found %d", len(foundLayers))
	}
	res.BinaryFile = foundLayers[0].Annotations["org.opencontainers.image.title"]

	//sourcePath := filepath.Join(tempDir.Path, fileName)

	// get the control file info
	foundLayers = ociinstaller.FindLayersForMediaType(layers, MediaTypeFdwControlLayer)
	if len(foundLayers) != 1 {
		return nil, fmt.Errorf("invalid image - image should contain 1 control file, found %d", len(foundLayers))
	}
	res.ControlFile = foundLayers[0].Annotations["org.opencontainers.image.title"]

	// get the sql file info
	foundLayers = ociinstaller.FindLayersForMediaType(layers, MediaTypeFdwSqlLayer)
	if len(foundLayers) != 1 {
		return nil, fmt.Errorf("invalid image - image should contain 1 SQL file, found %d", len(foundLayers))
	}
	res.SqlFile = foundLayers[0].Annotations["org.opencontainers.image.title"]

	// get the readme file info
	foundLayers = ociinstaller.FindLayersForMediaType(layers, MediaTypeFdwDocLayer)
	if len(foundLayers) > 0 {
		res.ReadmeFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	// get the license file info
	foundLayers = ociinstaller.FindLayersForMediaType(layers, MediaTypeFdwLicenseLayer)
	if len(foundLayers) > 0 {
		res.LicenseFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}
	return res, nil
}
