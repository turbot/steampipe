package ociinstaller

import (
	"context"
	"log"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/oras"
)

type ociDownloader struct{}

// NewOciDownloader creates and returns a ociDownloader instance
func NewOciDownloader() *ociDownloader {
	return new(ociDownloader)
}

/*
*

Pull downloads the image from the given `ref` to the supplied `destDir`

Returns

	imageDescription, configDescription, config, imageLayers, error

*
*/
func (o *ociDownloader) Pull(ctx context.Context, ref string, mediaTypes []string, destDir string) (*ocispec.Descriptor, []byte, []ocispec.Descriptor, error) {
	log.Println("[TRACE] ociDownloader.Pull:", "pulling", ref)
	fileStore := content.NewFile(destDir)
	defer fileStore.Close()

	layers := []ocispec.Descriptor{}

	hybridStore := newHybridStore(fileStore)
	pullOpts := []oras.CopyOpt{
		oras.WithAllowedMediaTypes(append(mediaTypes, MediaTypeConfig, MediaTypePluginConfig)),
		oras.WithPullEmptyNameAllowed(),
		oras.WithLayerDescriptors(func(d []ocispec.Descriptor) {
			layers = d
		}),
	}

	// An OCI Compliant registry is the source
	registry, err := content.NewRegistry(content.RegistryOptions{})
	if err != nil {
		return nil, nil, nil, err
	}

	desc, err := oras.Copy(ctx, registry, ref, WrappedHybridStore(hybridStore), ref, pullOpts...)
	if err != nil {
		return nil, nil, nil, err
	}
	_, configData, err := hybridStore.GetConfig()
	return &desc, configData, layers, err
}
