package ociinstaller

// This is required to get the config, a capability which may be added
// to oras as some point in the future.
// See https://github.com/deislabs/oras/issues/142#issuecomment-551735276

import (
	"context"
	"errors"
	"regexp"

	ocicontent "github.com/containerd/containerd/content"
	"github.com/containerd/containerd/remotes"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/target"
)

type hybridStore struct {
	remotes.Resolver
	*content.Memory
	*content.File
	config *ocispec.Descriptor
}

func newHybridStore(persist *content.File) *hybridStore {
	return &hybridStore{
		Memory: content.NewMemory(),
		File:   persist,
	}
}

func (s *hybridStore) GetConfig() (ocispec.Descriptor, []byte, error) {
	if s.config != nil {
		if desc, data, found := s.Memory.Get(*s.config); found {
			return desc, data, nil
		}
	}
	return ocispec.Descriptor{}, nil, errors.New("config not found")
}

// Writer begins or resumes the active writer identified by desc
func (s *hybridStore) Push(ctx context.Context, desc ocispec.Descriptor) (ocicontent.Writer, error) {
	pusher, err := s.File.Pusher(ctx, "")
	if err != nil {
		return nil, err
	}
	if isConfigMediaType(desc.MediaType) {
		s.config = &desc
		pusher, err = s.Memory.Pusher(ctx, "")
		if err != nil {
			return nil, err
		}
	}
	return pusher.Push(ctx, desc)
}

func isConfigMediaType(mediaType string) bool {
	// assume that any media type that ends with `.config.v?+json` is a config,
	// as well as the oci MediaTypeImageManifest and  MediaTypeImageIndex
	// 		ex: "application/vnd.turbot.steampipe.plugin.config.v1+json"
	matched, _ := regexp.MatchString(`.+\.config\.v\d\+json$`, mediaType)
	return matched || mediaType == ocispec.MediaTypeImageManifest || mediaType == ocispec.MediaTypeImageIndex
}

type pushTarget struct {
	target.Target
	hStore *hybridStore
}

func (c pushTarget) Resolve(ctx context.Context, ref string) (name string, desc ocispec.Descriptor, err error) {
	return "", ocispec.Descriptor{}, errors.New("not implemented")
}

func (c pushTarget) Fetcher(ctx context.Context, ref string) (remotes.Fetcher, error) {
	return nil, errors.New("not implemented")
}

func (c pushTarget) Pusher(ctx context.Context, ref string) (remotes.Pusher, error) {
	return c.hStore, nil
}

func WrappedHybridStore(h *hybridStore) target.Target {
	return pushTarget{
		hStore: h,
	}
}
