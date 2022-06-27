package ociinstaller

// This is required to get the config, a capability which may be added
// to oras as some point in the future.
// See https://github.com/deislabs/oras/issues/142#issuecomment-551735276

import (
	"context"
	"errors"
	"regexp"

	ocicontent "github.com/containerd/containerd/content"
	"github.com/deislabs/oras/pkg/content"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type hybridStore struct {
	*content.Memorystore
	ingester ocicontent.Ingester
	config   *ocispec.Descriptor
}

func newHybridStore(ingester ocicontent.Ingester) *hybridStore {
	return &hybridStore{
		Memorystore: content.NewMemoryStore(),
		ingester:    ingester,
	}
}

func (s *hybridStore) GetConfig() (ocispec.Descriptor, []byte, error) {
	if s.config != nil {
		if desc, data, found := s.Memorystore.Get(*s.config); found {
			return desc, data, nil
		}
	}
	return ocispec.Descriptor{}, nil, errors.New("config not found")
}

// Writer begins or resumes the active writer identified by desc
func (s *hybridStore) Writer(ctx context.Context, opts ...ocicontent.WriterOpt) (ocicontent.Writer, error) {
	var wOpts ocicontent.WriterOpts
	for _, opt := range opts {
		if err := opt(&wOpts); err != nil {
			return nil, err
		}
	}

	if isConfigMediaType(wOpts.Desc.MediaType) {
		s.config = &wOpts.Desc
		return s.Memorystore.Writer(ctx, opts...)

	}

	return s.ingester.Writer(ctx, opts...)
}

func isConfigMediaType(mediaType string) bool {
	// assume that any media type that ends with `.config.v?+json` is a config,
	// as well as the oci MediaTypeImageManifest and  MediaTypeImageIndex
	// 		ex: "application/vnd.turbot.steampipe.plugin.config.v1+json"
	matched, _ := regexp.MatchString(`.+\.config\.v\d\+json$`, mediaType)
	return matched || mediaType == ocispec.MediaTypeImageManifest || mediaType == ocispec.MediaTypeImageIndex
}
