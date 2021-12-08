package ociinstaller

import (
	"context"
	"errors"
	"fmt"

	"github.com/containerd/containerd/remotes"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type SteampipeImage struct {
	OCIDescriptor *ocispec.Descriptor
	ImageRef      string
	Config        *config
	Plugin        *PluginImage
	Database      *DbImage
	Fdw           *HubImage
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

func (o *ociDownloader) newSteampipeImage() *SteampipeImage {
	SteampipeImage := &SteampipeImage{
		resolver: &o.resolver,
	}
	o.Images = append(o.Images, SteampipeImage)
	return SteampipeImage
}

func (o *ociDownloader) Download(ctx context.Context, ref string, imageType string, destDir string) (*SteampipeImage, error) {
	var mediaTypes []string
	Image := o.newSteampipeImage()
	Image.ImageRef = ref

	mediaTypes = append(mediaTypes, MediaTypeForPlatform(imageType))
	mediaTypes = append(mediaTypes, SharedMediaTypes(imageType)...)
	mediaTypes = append(mediaTypes, ConfigMediaTypes()...)

	// Download the files
	imageDesc, _, configBytes, layers, err := o.Pull(ctx, ref, mediaTypes, destDir)
	if err != nil {
		return nil, err
	}

	Image.OCIDescriptor = imageDesc
	Image.Config, err = newSteampipeImageConfig(configBytes)
	if err != nil {
		return nil, errors.New("Invalid image - missing $config")
	}

	// Get the metadata
	switch imageType {
	case "db":
		Image.Database, err = getDBImageData(layers)
	case "fdw":
		Image.Fdw, err = getHubImageData(layers)
	case "plugin":
		Image.Plugin, err = getPluginImageData(layers)

	default:
		return nil, errors.New("Invalid Type - Image types are: plugin, db, fdw")
	}

	if err != nil {
		return nil, err
	}
	return Image, nil
}

func getDBImageData(layers []ocispec.Descriptor) (*DbImage, error) {
	var DbImage DbImage

	// get the binary jar file
	foundLayers := findLayersForMediaType(layers, MediaTypeForPlatform("db"))
	if len(foundLayers) != 1 {
		return nil, fmt.Errorf("Invalid Image - Image should contain 1 installation file per platform, found %d", len(foundLayers))
	}
	DbImage.ArchiveDir = foundLayers[0].Annotations["org.opencontainers.image.title"]

	// get the readme file info
	foundLayers = findLayersForMediaType(layers, MediaTypeDbDocLayer)
	if len(foundLayers) > 0 {
		DbImage.ReadmeFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	// get the license file info
	foundLayers = findLayersForMediaType(layers, MediaTypeDbLicenseLayer)
	if len(foundLayers) > 0 {
		DbImage.LicenseFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}
	return &DbImage, nil
}

func getHubImageData(layers []ocispec.Descriptor) (*HubImage, error) {
	var HubImage HubImage

	// get the binary (steampipe-postgres-fdw.so) info
	foundLayers := findLayersForMediaType(layers, MediaTypeForPlatform("fdw"))
	if len(foundLayers) != 1 {
		return nil, fmt.Errorf("Invalid Image - Image should contain 1 binary file per platform, found %d", len(foundLayers))
	}
	HubImage.BinaryFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	//sourcePath := filepath.Join(tempDir.Path, fileName)

	// get the control file info
	foundLayers = findLayersForMediaType(layers, MediaTypeFdwControlLayer)
	if len(foundLayers) != 1 {
		return nil, fmt.Errorf("Invalid Image - Image should contain 1 control file, found %d", len(foundLayers))
	}
	HubImage.ControlFile = foundLayers[0].Annotations["org.opencontainers.image.title"]

	// get the sql file info
	foundLayers = findLayersForMediaType(layers, MediaTypeFdwSqlLayer)
	if len(foundLayers) != 1 {
		return nil, fmt.Errorf("Invalid Image - Image should contain 1 SQL file, found %d", len(foundLayers))
	}
	HubImage.SqlFile = foundLayers[0].Annotations["org.opencontainers.image.title"]

	// get the readme file info
	foundLayers = findLayersForMediaType(layers, MediaTypeFdwDocLayer)
	if len(foundLayers) > 0 {
		HubImage.ReadmeFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	// get the license file info
	foundLayers = findLayersForMediaType(layers, MediaTypeFdwLicenseLayer)
	if len(foundLayers) > 0 {
		HubImage.LicenseFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}
	return &HubImage, nil
}

func getPluginImageData(layers []ocispec.Descriptor) (*PluginImage, error) {
	var PluginImage PluginImage

	// get the binary plugin file info
	foundLayers := findLayersForMediaType(layers, MediaTypeForPlatform("plugin"))
	if len(foundLayers) != 1 {
		return nil, fmt.Errorf("Invalid Image - Image should contain 1 binary file per platform, found %d", len(foundLayers))
	}
	PluginImage.BinaryFile = foundLayers[0].Annotations["org.opencontainers.image.title"]

	// get the docs dir
	foundLayers = findLayersForMediaType(layers, MediaTypePluginDocsLayer)
	if len(foundLayers) > 0 {
		PluginImage.DocsDir = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	// get the .spc config / connections file dir
	foundLayers = findLayersForMediaType(layers, MediaTypePluginSpcLayer)
	if len(foundLayers) > 0 {
		PluginImage.ConfigFileDir = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	// get the license file info
	foundLayers = findLayersForMediaType(layers, MediaTypePluginLicenseLayer)
	if len(foundLayers) > 0 {
		PluginImage.LicenseFile = foundLayers[0].Annotations["org.opencontainers.image.title"]
	}

	return &PluginImage, nil
}

func findLayersForMediaType(layers []ocispec.Descriptor, mediaType string) []ocispec.Descriptor {
	var matchedLayers []ocispec.Descriptor

	for _, layer := range layers {
		if layer.MediaType == mediaType {
			matchedLayers = append(matchedLayers, layer)
		}
	}
	return matchedLayers
}
