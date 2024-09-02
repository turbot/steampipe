package ociinstaller

import (
	"fmt"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/turbot/pipe-fittings/ociinstaller"
)

const (
	ImageTypeDatabase ociinstaller.ImageType = "db"
	ImageTypeFdw      ociinstaller.ImageType = "fdw"
	ImageTypeAssets   ociinstaller.ImageType = "assets"
)

type DbImage struct {
	ArchiveDir  string
	ReadmeFile  string
	LicenseFile string
}

func (s *DbImage) Type() ociinstaller.ImageType {
	return ImageTypeDatabase
}

type FdwImage struct {
	BinaryFile  string
	ReadmeFile  string
	LicenseFile string
	ControlFile string
	SqlFile     string
}

func (s *FdwImage) Type() ociinstaller.ImageType {
	return ImageTypeFdw
}

type AssetsImage struct {
	ReportUI string
}

func (s *AssetsImage) Type() ociinstaller.ImageType {
	return ImageTypeAssets
}

func getAssetImageData(layers []ocispec.Descriptor) (*AssetsImage, error) {
	var assetImage AssetsImage

	// get the report dir
	foundLayers := ociinstaller.FindLayersForMediaType(layers, MediaTypeAssetReportLayer)
	if len(foundLayers) > 0 {
		assetImage.ReportUI = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	return &assetImage, nil
}

func getDBImageData(layers []ocispec.Descriptor) (*DbImage, error) {
	res := &DbImage{}

	// get the binary jar file
	mediaType, err := MediaTypeForPlatform("db")
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

func getFdwImageData(layers []ocispec.Descriptor) (*FdwImage, error) {
	res := &FdwImage{}

	// get the binary (steampipe-postgres-fdw.so) info
	mediaType, err := MediaTypeForPlatform("fdw")
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
