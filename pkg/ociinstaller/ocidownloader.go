package ociinstaller

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	credentials "github.com/oras-project/oras-credentials-go"
	"github.com/sirupsen/logrus"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
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

/*
Pull downloads the image from the given `ref` to the supplied `destDir`

Returns

	imageDescription, configDescription, config, imageLayers, error
*/
func (o *ociDownloader) Pull(ctx context.Context, ref string, mediaTypes []string, destDir string) (*ocispec.Descriptor, *ocispec.Descriptor, []byte, []ocispec.Descriptor, error) {
	split := strings.Split(ref, ":")
	tag := split[len(split)-1]
	log.Println("[TRACE] ociDownloader.Pull:", "preparing to pull ref", ref, "tag", tag, "destDir", destDir)

	// Create the target file store
	memoryStore := memory.New()
	fileStore, err := file.NewWithFallbackStorage(destDir, memoryStore)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer fileStore.Close()

	// Connect to the remote repository
	repo, err := remote.NewRepository(ref)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Get credentials from the docker credentials store
	storeOpts := credentials.StoreOptions{}
	credStore, err := credentials.NewStoreFromDocker(storeOpts)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Prepare the auth client for the registry and credential store
	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.DefaultCache,
		Credential: credentials.Credential(credStore), // Use the credential store
	}

	// Copy from the remote repository to the file store
	log.Println("[TRACE] ociDownloader.Pull:", "pulling...")

	copyOpt := oras.DefaultCopyOptions
	manifestDescriptor, err := oras.Copy(ctx, repo, tag, fileStore, tag, copyOpt)
	if err != nil {
		log.Println("[TRACE] ociDownloader.Pull:", "failed to pull", ref, err)
		return nil, nil, nil, nil, err
	}
	log.Println("[TRACE] ociDownloader.Pull:", "manifest", manifestDescriptor.Digest, manifestDescriptor.MediaType)

	// FIXME: this seems redundant as oras.Copy() already downloads all artifacts, but that's the only I found
	// to access the manifest config. Also, it shouldn't be an issue as files are not re-downloaded.
	manifestJson, err := content.FetchAll(ctx, fileStore, manifestDescriptor)
	if err != nil {
		log.Println("[TRACE] ociDownloader.Pull:", "failed to fetch manifest", manifestDescriptor)
		return nil, nil, nil, nil, err
	}
	log.Println("[TRACE] ociDownloader.Pull:", "manifest content", string(manifestJson))

	// Parse the fetched manifest
	var manifest ocispec.Manifest
	err = json.Unmarshal(manifestJson, &manifest)
	if err != nil {
		log.Println("[TRACE] ociDownloader.Pull:", "failed to unmarshall manifest", manifestJson)
		return nil, nil, nil, nil, err
	}

	// Fetch the config from the file store
	configData, err := content.FetchAll(ctx, fileStore, manifest.Config)
	if err != nil {
		log.Println("[TRACE] ociDownloader.Pull:", "failed to fetch config", manifest.Config.MediaType, err)
		return nil, nil, nil, nil, err
	}
	log.Println("[TRACE] ociDownloader.Pull:", "config", string(configData))

	return &manifestDescriptor, &manifest.Config, configData, manifest.Layers, err
}
