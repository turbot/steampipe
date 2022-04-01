package ociinstaller

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/containerd/containerd/remotes"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type SteampipeImage struct {
	OCIDescriptor *ocispec.Descriptor
	ImageRef      *SteampipeImageRef
	Config        *config
	Plugin        *PluginImage
	Database      *DbImage
	Fdw           *HubImage
	Assets        *AssetsImage
	resolver      *remotes.Resolver
}

type PluginImage struct {
	BinaryFile    string
	DocsDir       string
	ConfigFileDir string
	LicenseFile   string
}

type DbImage struct {
	ArchiveDir  string
	ReadmeFile  string
	LicenseFile string
}
type HubImage struct {
	BinaryFile  string
	ReadmeFile  string
	LicenseFile string
	ControlFile string
	SqlFile     string
}
type AssetsImage struct {
	ReportUI string
}

func (o *ociDownloader) newSteampipeImage() *SteampipeImage {
	SteampipeImage := &SteampipeImage{
		resolver: &o.resolver,
	}
	o.Images = append(o.Images, SteampipeImage)
	return SteampipeImage
}

type ImageType string

const (
	ImageTypeDatabase ImageType = "db"
	ImageTypeFdw      ImageType = "fdw"
	ImageTypeAssets   ImageType = "assets"
	ImageTypePlugin   ImageType = "plugin"
)

func (o *ociDownloader) Download(ctx context.Context, ref *SteampipeImageRef, imageType ImageType, destDir string) (*SteampipeImage, error) {
	var mediaTypes []string
	Image := o.newSteampipeImage()
	Image.ImageRef = ref

	mediaTypes = append(mediaTypes, MediaTypeForPlatform(imageType)...)
	mediaTypes = append(mediaTypes, SharedMediaTypes(imageType)...)
	mediaTypes = append(mediaTypes, ConfigMediaTypes()...)

	log.Println("[TRACE] ociDownloader.Download:", "downloading", ref.ActualImageRef())

	// Download the files
	imageDesc, _, configBytes, layers, err := o.Pull(ctx, ref.ActualImageRef(), mediaTypes, destDir)
	if err != nil {
		return nil, err
	}

	Image.OCIDescriptor = imageDesc
	Image.Config, err = newSteampipeImageConfig(configBytes)
	if err != nil {
		return nil, errors.New("invalid image - missing $config")
	}

	// Get the metadata
	switch imageType {
	case ImageTypeDatabase:
		Image.Database, err = getDBImageData(layers)
	case ImageTypeFdw:
		Image.Fdw, err = getFdwImageData(layers)
	case ImageTypePlugin:
		Image.Plugin, err = getPluginImageData(layers)
	case ImageTypeAssets:
		Image.Assets, err = getAssetImageData(layers)

	default:
		return nil, errors.New("invalid Type - Image types are: plugin, db, fdw")
	}

	if err != nil {
		return nil, err
	}
	return Image, nil
}

func getAssetImageData(layers []ocispec.Descriptor) (*AssetsImage, error) {
	var assetImage AssetsImage

	// get the report dir
	foundLayers := findLayersForMediaType(layers, MediaTypeAssetReportLayer)
	if len(foundLayers) > 0 {
		assetImage.ReportUI = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	return &assetImage, nil
}

func getDBImageData(layers []ocispec.Descriptor) (*DbImage, error) {
	res := &DbImage{}

	// get the binary jar file
	foundLayers := findLayersForMediaType(layers, MediaTypeForPlatform("db")[0])
	if len(foundLayers) != 1 {
		return nil, fmt.Errorf("invalid Image - Image should contain 1 installation file per platform, found %d", len(foundLayers))
	}
	res.ArchiveDir = foundLayers[0].Annotations["org.opencontainers.image.title"]

	// get the readme file info
	foundLayers = findLayersForMediaType(layers, MediaTypeDbDocLayer)
	if len(foundLayers) > 0 {
		res.ReadmeFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	// get the license file info
	foundLayers = findLayersForMediaType(layers, MediaTypeDbLicenseLayer)
	if len(foundLayers) > 0 {
		res.LicenseFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}
	return res, nil
}

func getFdwImageData(layers []ocispec.Descriptor) (*HubImage, error) {
	res := &HubImage{}

	// get the binary (steampipe-postgres-fdw.so) info
	foundLayers := findLayersForMediaType(layers, MediaTypeForPlatform("fdw")[0])
	if len(foundLayers) != 1 {
		return nil, fmt.Errorf("invalid image - Image should contain 1 binary file per platform, found %d", len(foundLayers))
	}
	res.BinaryFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	//sourcePath := filepath.Join(tempDir.Path, fileName)

	// get the control file info
	foundLayers = findLayersForMediaType(layers, MediaTypeFdwControlLayer)
	if len(foundLayers) != 1 {
		return nil, fmt.Errorf("invalid image - Image should contain 1 control file, found %d", len(foundLayers))
	}
	res.ControlFile = foundLayers[0].Annotations["org.opencontainers.image.title"]

	// get the sql file info
	foundLayers = findLayersForMediaType(layers, MediaTypeFdwSqlLayer)
	if len(foundLayers) != 1 {
		return nil, fmt.Errorf("invalid image - Image should contain 1 SQL file, found %d", len(foundLayers))
	}
	res.SqlFile = foundLayers[0].Annotations["org.opencontainers.image.title"]

	// get the readme file info
	foundLayers = findLayersForMediaType(layers, MediaTypeFdwDocLayer)
	if len(foundLayers) > 0 {
		res.ReadmeFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	// get the license file info
	foundLayers = findLayersForMediaType(layers, MediaTypeFdwLicenseLayer)
	if len(foundLayers) > 0 {
		res.LicenseFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}
	return res, nil
}

func getPluginImageData(layers []ocispec.Descriptor) (*PluginImage, error) {
	res := &PluginImage{}
	var foundLayers []ocispec.Descriptor
	// get the binary plugin file info
	for _, mediaType := range MediaTypeForPlatform("plugin") {
		foundLayers = findLayersForMediaType(layers, mediaType)
		if len(foundLayers) != 1 {
			log.Println("[TRACE] could not find data for", mediaType)
			log.Println("[TRACE] falling back to the next one, if any")
			continue
		}
		res.BinaryFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}
	if len(res.BinaryFile) == 0 {
		return nil, fmt.Errorf("invalid Image - Image should contain 1 binary file per platform, found %d", len(foundLayers))
	}

	// get the docs dir
	foundLayers = findLayersForMediaType(layers, MediaTypePluginDocsLayer)
	if len(foundLayers) > 0 {
		res.DocsDir = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	// get the .spc config / connections file dir
	foundLayers = findLayersForMediaType(layers, MediaTypePluginSpcLayer)
	if len(foundLayers) > 0 {
		res.ConfigFileDir = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	// get the license file info
	foundLayers = findLayersForMediaType(layers, MediaTypePluginLicenseLayer)
	if len(foundLayers) > 0 {
		res.LicenseFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	return res, nil
}

func findLayersForMediaType(layers []ocispec.Descriptor, mediaType string) []ocispec.Descriptor {
	log.Println("[TRACE] looking for", mediaType)
	var matchedLayers []ocispec.Descriptor

	for _, layer := range layers {
		if layer.MediaType == mediaType {
			matchedLayers = append(matchedLayers, layer)
		}
	}
	return matchedLayers
}
