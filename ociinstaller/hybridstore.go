package ociinstaller

import (
	"context"
	"io"

	"github.com/containerd/containerd/content"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	orascontent "oras.land/oras-go/pkg/content"
)

// LayeredStore has a writable cache on the top and a provider at the bottom
type LayeredStore struct {
	*orascontent.Memory
	provider orascontent.Store
}

// NewLayeredStore create a new layered store
func NewLayeredStore(provider orascontent.Store) LayeredStore {
	return LayeredStore{
		Memory:   orascontent.NewMemory(),
		provider: provider,
	}
}

// ReaderAt reads from the cache first and fallback to the provider
func (s LayeredStore) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	if isConfigMediaType(desc.MediaType) {
		return s.Memory.Fetch(ctx, desc)
	}
	return s.provider.Fetch(ctx, desc)
}

// ReaderAt reads from the cache first and fallback to the provider
func (s LayeredStore) Push(ctx context.Context, d ocispec.Descriptor) (content.Writer, error) {
	p, _ := s.Memory.Pusher(ctx, "")
	p.Push(ctx, d)
	return s.provider.Push(ctx, ref)
}

// This is required to get the config, a capability which may be added
// to oras as some point in the future.
// See https://github.com/deislabs/oras/issues/142#issuecomment-551735276

// import (
// 	"context"
// 	"errors"
// 	"io"
// 	"regexp"

// 	ocicontent "github.com/containerd/containerd/content"
// 	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
// 	"oras.land/oras-go/pkg/content"
// )

// type hybridStore struct {
// 	*content.Memory
// 	fileStore *content.File
// 	ingester  ocicontent.Ingester
// 	config    *ocispec.Descriptor
// }

// func newHybridStore(fileStore *content.File) *hybridStore {
// 	h := &hybridStore{
// 		Memory:    content.NewMemory(),
// 		fileStore: fileStore,
// 	}
// 	return h
// }

// func (s *hybridStore) GetConfig() (ocispec.Descriptor, []byte, error) {
// 	if s.config != nil {
// 		if desc, data, found := s.Memory.Get(*s.config); found {
// 			return desc, data, nil
// 		}
// 	}
// 	return ocispec.Descriptor{}, nil, errors.New("config not found")
// }

// func (s *hybridStore) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
// 	if isConfigMediaType(desc.MediaType) {
// 		s.config = &desc
// 		s.Memory.Add()
// 		return s.Memory.Fetch(ctx, desc)
// 	}
// 	return s.fileStore.Fetch(ctx, desc)
// }

// func isConfigMediaType(mediaType string) bool {
// 	// assume that any media type that ends with `.config.v?+json` is a config,
// 	// as well as the oci MediaTypeImageManifest and  MediaTypeImageIndex
// 	// 		ex: "application/vnd.turbot.steampipe.plugin.config.v1+json"
// 	matched, _ := regexp.MatchString(`.+\.config\.v\d\+json$`, mediaType)
// 	return matched || mediaType == ocispec.MediaTypeImageManifest || mediaType == ocispec.MediaTypeImageIndex
// }
