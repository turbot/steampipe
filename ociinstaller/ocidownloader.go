package ociinstaller

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/containerd/containerd/remotes"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"github.com/turbot/steampipe/utils"

	orascontent "oras.land/oras-go/pkg/content"
	orasgo "oras.land/oras-go/pkg/oras"
)

type ociDownloader struct {
	resolver remotes.Resolver
	Images   []*SteampipeImage
}

// NewOciDownloader creates and returns a ociDownloader instance
func NewOciDownloader() *ociDownloader {
	// oras uses containerd, which uses logrus and is set up to log
	// warning and above.  Set to ErrrLevel to get rid of unwanted error message
	logrus.SetLevel(logrus.ErrorLevel)
	return &ociDownloader{
		// resolver: docker.NewResolver(docker.ResolverOptions{}),
	}
}

/**

Pull downloads the image from the given `ref` to the supplied `destDir`

Returns
	imageDescription, configDescription, config, imageLayers, error

**/
func (o *ociDownloader) Pull(ctx context.Context, ref string, mediaTypes []string, destDir string) (*ocispec.Descriptor, *ocispec.Descriptor, []byte, []ocispec.Descriptor, error) {
	fileStore := orascontent.NewFile(destDir)
	defer fileStore.Close()

	registryTarget, _ := orascontent.NewRegistry(orascontent.RegistryOptions{Insecure: false})

	layerStore := []ocispec.Descriptor{}

	var (
		configDesc ocispec.Descriptor
		configData []byte
	)

	pullOpts := []orasgo.CopyOpt{
		orasgo.WithAllowedMediaTypes(append(mediaTypes, MediaTypeConfig, MediaTypePluginConfig)),
		orasgo.WithPullEmptyNameAllowed(),
		orasgo.WithLayerDescriptors(func(d []ocispec.Descriptor) {
			layerStore = append(layerStore, d...)
			utils.DebugDumpJSON("layers:", layerStore)
		}),
		orasgo.WithRootManifest(func(b []byte) {
			config := map[string]interface{}{}
			json.Unmarshal(b, &config)
			utils.DebugDumpJSON("manifest:", config)
			configData = b
		}),
	}

	desc, err := orasgo.Copy(ctx, registryTarget, ref, fileStore, "", pullOpts...)
	if err != nil {
		return &desc, nil, nil, nil, err
	}
	return &desc, &configDesc, configData, layerStore, err
}

func isConfigMediaType(mediaType string) bool {
	// assume that any media type that ends with `.config.v?+json` is a config,
	// as well as the oci MediaTypeImageManifest and  MediaTypeImageIndex
	// 		ex: "application/vnd.turbot.steampipe.plugin.config.v1+json"
	matched, _ := regexp.MatchString(`.+\.config\.v\d\+json$`, mediaType)
	return matched || mediaType == ocispec.MediaTypeImageManifest || mediaType == ocispec.MediaTypeImageIndex
}
