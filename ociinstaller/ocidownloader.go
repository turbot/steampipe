package ociinstaller

import (
	"context"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
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
		resolver: docker.NewResolver(docker.ResolverOptions{}),
	}
}

/**

Pull downloads the image from the given `ref` to the supplied `destDir`

Returns
	imageDescription, configDescription, config, imageLayers, error

**/
func (o *ociDownloader) Pull(ctx context.Context, ref string, mediaTypes []string, destDir string) (*ocispec.Descriptor, *ocispec.Descriptor, []byte, []ocispec.Descriptor, error) {
	fileStore := content.NewFileStore(destDir)
	defer fileStore.Close()

	hybridStore := newHybridStore(fileStore)
	pullOpts := []oras.PullOpt{
		oras.WithAllowedMediaTypes(append(mediaTypes, MediaTypeConfig, MediaTypePluginConfig)),
		oras.WithPullEmptyNameAllowed(),
	}
	desc, layers, err := oras.Pull(ctx, o.resolver, ref, hybridStore, pullOpts...)
	if err != nil {
		return &desc, nil, nil, layers, err
	}
	configDesc, configData, err := hybridStore.GetConfig()
	return &desc, &configDesc, configData, layers, err
}
