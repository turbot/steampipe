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
	context  context.Context
	Images   []*SteampipeImage
}

// NewOciDownloader :: creats and return a ociDownloader instance
func NewOciDownloader(ctx context.Context) *ociDownloader {
	// oras uses containerd, which uses logrus and is set up to log
	// warning and above.  Set to ErrrLevel to get rid of unwanted error message
	logrus.SetLevel(logrus.ErrorLevel)
	return &ociDownloader{
		resolver: docker.NewResolver(docker.ResolverOptions{}),
		context:  ctx,
	}
}

func (o *ociDownloader) Pull(ref string, mediaTypes []string, destDir string) (ocispec.Descriptor, []byte, []ocispec.Descriptor, error) {
	fileStore := content.NewFileStore(destDir)
	defer fileStore.Close()

	hybridStore := newHybridStore(fileStore)
	pullOpts := []oras.PullOpt{
		oras.WithAllowedMediaTypes(append(mediaTypes, MediaTypeConfig, MediaTypePluginConfig)),
		oras.WithPullEmptyNameAllowed(),
	}
	desc, layers, err := oras.Pull(o.context, o.resolver, ref, hybridStore, pullOpts...)
	if err != nil {
		return desc, nil, layers, err
	}

	configDesc, configData, err := hybridStore.GetConfig()
	return configDesc, configData, layers, err
}
